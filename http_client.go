package base

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jie123108/glog"
)

func init() {
	//fmt.Printf("############# Build Time: %s #############\n", BuildTime)
}

const (
	ERR_ARGS_INVALID string = "ERR_ARGS_INVALID"
	ERR_SERVER_ERROR string = "ERR_SERVER_ERROR"
)

type ReqStats struct {
	All time.Duration //请求总时间。
}

type Resp struct {
	Error      error       // 出错信息。
	RawBody    []byte      // Http返回的原始内容。
	StatusCode int         // Http响应吗
	Headers    http.Header // HTTP响应头
	ReqDebug   string      // 请求的DEBUG串(curl格式)
	Stats      ReqStats    //请求时间统计。
}

type OkJson struct {
	Resp
	Ok     bool                   `json:"ok"`
	Reason string                 `json:"reason"`
	Data   map[string]interface{} `json:"data"`
	Cached string                 //数据是否是缓存中获取 hit, miss
}

func (this *OkJson) Body() string {
	return string(this.RawBody)
}

func headerstr(headers map[string]string) string {
	if headers == nil {
		return ""
	}

	lines := make([]string, 4)
	for k, v := range headers {
		if k != "User-Agent" {
			lines = append(lines, "-H'"+k+": "+v+"'")
		}
	}

	return strings.Join(lines, " ")
}

var g_proxy_url string

func HttpHeaderConv(src_headers http.Header) (headers map[string]string) {
	headers = make(map[string]string)
	if src_headers != nil {
		for k, varr := range src_headers {
			headers[k] = varr[0]
		}
	}
	return
}

// proxyURL : "http://" + p.AppID + ":" + p.AppSecret + "@" + ProxyServer
func SetProxyURL(proxyURL string) {
	g_proxy_url = proxyURL
}

func is_text_context(content_type string, headers map[string]string) bool {
	ok := content_type == "" || strings.HasPrefix(content_type, "text") ||
		strings.HasPrefix(content_type, "application/json") ||
		strings.HasPrefix(content_type, "application/x-www-form-urlencoded;charset=utf-8") ||
		headers["X-Body-Is-Text"] == "1"
	return ok
}

func HttpReqDebug(method, uri string, body []byte, headers map[string]string, max_body_len int) string {
	var req_debug string
	if method == "PUT" || method == "POST" {
		var debug_body string
		content_type := headers["Content-Type"]
		if is_text_context(content_type, headers) {
			if max_body_len == 0 || len(body) < max_body_len {
				debug_body = string(body)
			} else {
				debug_body = string(body[0:max_body_len])
			}
		} else {
			debug_body = "[[not text body: " + content_type + "]]"
		}
		req_debug = "curl -v -X " + method + " " + headerstr(headers) + " '" + uri + "' -d '" + debug_body + "' "
	} else {
		req_debug = "curl -v -X " + method + " " + headerstr(headers) + " '" + uri + "' "
	}
	return req_debug
}

//支持原生调用，wangyanglong@nicefilm.com
//外部需要自己回收资源 defer resp.body.Close()
//https://golang.org/src/net/http/client.go
// The Client's Transport typically has internal state (cached TCP
// connections), so Clients should be reused instead of created as
// needed. Clients are safe for concurrent use by multiple goroutines.
var g_default_client *http.Client
var onceDefaultClient sync.Once
var onceProxyClient sync.Once

type proxyClientSet struct {
	sync.RWMutex
	clientMap map[string]*http.Client
}

var g_proxy_client *proxyClientSet

func newHttpClient(transport *http.Transport) *http.Client {
	timeout := time.Minute
	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}
	return client
}

const (
	MaxIdleConnsPerHost = 15
	MaxIdleConns        = 500
	DefTimeOut          = 10 * time.Second
)

func newHttpTransport() *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          MaxIdleConns,
		MaxIdleConnsPerHost:   MaxIdleConnsPerHost,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 5 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return transport
}

func getDefaultHttpClient() (client *http.Client, err error) {
	onceDefaultClient.Do(func() {
		g_default_client = newHttpClient(newHttpTransport())
	})
	if g_default_client != nil {
		client = g_default_client
	} else {
		err = errors.New("get default http client error,init failed")
	}
	return
}

func getProxyHttpClient(proxy string) (client *http.Client, err error) {
	onceProxyClient.Do(func() {
		g_proxy_client = new(proxyClientSet)
		g_proxy_client.clientMap = make(map[string]*http.Client)
	})
	g_proxy_client.Lock()
	defer g_proxy_client.Unlock()

	client, ok := g_proxy_client.clientMap[proxy]
	if ok && client != nil {
		return
	}
	transport := newHttpTransport()
	proxyURL, _ := url.Parse(proxy)
	transport.Proxy = http.ProxyURL(proxyURL)
	client = newHttpClient(transport)
	g_proxy_client.clientMap[proxy] = client
	return
}

func Http_req(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) (*http.Response, error, string) {
	client, err := getDefaultHttpClient()
	if err != nil {
		return nil, err, ""
	}
	not_use_proxy := headers["X-Not-Use-Proxy"] == "true"
	if not_use_proxy {
		delete(headers, "X-Not-Use-Proxy")
	}
	if !not_use_proxy && g_proxy_url != "" {
		client, err = getProxyHttpClient(g_proxy_url)
		if err != nil {
			return nil, err, ""
		}
	}
	req, err, _ := FormatHttpRequest(method, uri, args, headers, body)
	if err != nil {
		return nil, err, ""
	}
	url_with_args := req.URL.String()
	req_debug := HttpReqDebug(method, url_with_args, body, headers, 1024)

	//glog.Infof("REQUEST [ %s ] timeout: %v", req_debug, timeout)
	//begin := time.Now()
	client.Timeout = timeout    // add by wlj
	resp, err := client.Do(req) //发送
	// cost := time.Now().Sub(begin)
	// code := -1
	// if resp != nil {
	// 	code = resp.StatusCode
	// }
	// glog.Infof("REQUEST [ %s ] status: %d, recv header cost: %v", req_debug, code, cost)

	return resp, err, req_debug
}

func FormatHttpRequest(method, uri string, args, headers map[string]string, body []byte) (req *http.Request, err error, reason string) {
	req, err = http.NewRequest(method, uri, bytes.NewReader(body))
	if err != nil {
		glog.Errorf("NewRequest(method:%s, uri: '%s') failed! err: %v", method, uri, err)
		return nil, err, ""
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	q := req.URL.Query()
	for key, value := range args {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	req.Host = headers["Host"]
	return req, nil, ""
}

func write_debug(begin time.Time, res *Resp, body_len *int) {
	cost := time.Now().Sub(begin)
	res.Stats.All = cost
	seconds := cost.Seconds()
	kbps := float64(0)
	if seconds > 0 && *body_len > 0 {
		kbps = float64(*body_len) / float64(1024) / seconds
	}
	glog.Infof("REQUEST [ %s ] status: %d, body_len: %d, cost: %v, speed: %.3f kb/s", res.ReqDebug, res.StatusCode, *body_len, cost, kbps)
}

func write_debug_ok_json(begin time.Time, res *OkJson, body_len *int) {
	cost := time.Now().Sub(begin)
	res.Stats.All = cost
	seconds := cost.Seconds()
	kbps := float64(0)
	if seconds > 0 && *body_len > 0 {
		kbps = float64(*body_len) / float64(1024) / seconds
	}
	glog.Infof("REQUEST [ %s ] status: %d, body_len: %d, cost: %v, speed: %.3f kb/s", res.ReqDebug, res.StatusCode, *body_len, cost, kbps)
}

func HttpReqInternal(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) *Resp {
	res := &Resp{}
	begin := time.Now()
	body_len := 0

	resp, err, req_debug := Http_req(method, uri, body, args, headers, timeout)
	res.ReqDebug = req_debug
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close() //一定要关闭resp.Body
	}
	defer write_debug(begin, res, &body_len)
	if err != nil {
		glog.Errorf("###### err: %v", err)
		res.Error = err
		res.StatusCode = 500
		return res
	}

	res.Headers = resp.Header
	// defer resp.Body.Close() //一定要关闭resp.Body
	data, err := ioutil.ReadAll(resp.Body)
	if data != nil {
		body_len = len(data)
	}
	if err != nil {
		glog.Errorf("REQUEST [ %s ] Read Body Failed! err: %v", req_debug, err)
		res.StatusCode = 500
		res.Error = err
		return res
	}
	res.RawBody = data
	res.StatusCode = resp.StatusCode
	content_length := 0
	strContentLengths := res.Headers["Content-Length"]
	if len(strContentLengths) > 0 {
		content_length, _ = strconv.Atoi(strContentLengths[0])
		if content_length > 0 && len(data) != content_length {
			res.StatusCode = 500
			glog.Errorf("REQUEST [ %s ] Content-Length: %d, len(body): %d", res.ReqDebug, content_length, len(data))
			res.Error = fmt.Errorf("Length of Body is Invalid")
			return res
		}
	}
	if err != nil {
		glog.Errorf("REQUEST [ %s ] Read Body Failed! body-len: %d err: %v", req_debug, content_length, err)
		res.StatusCode = 500
		res.Error = err
		return res
	}

	return res
}

func HttpGet(uri string, headers map[string]string, timeout time.Duration) *Resp {
	return HttpReqInternal("GET", uri, nil, nil, headers, timeout)
}

func HttpPost(uri string, body []byte, headers map[string]string, timeout time.Duration) *Resp {
	return HttpReqInternal("POST", uri, body, nil, headers, timeout)
}

func OkJsonParse(res *OkJson) *OkJson {
	decoder := json.NewDecoder(bytes.NewBuffer(res.RawBody))
	decoder.UseNumber()
	err := decoder.Decode(&res)
	if err != nil {
		glog.Errorf("Invalid json [%s] err: %v", string(res.RawBody), err)
		res.Error = err
		res.Reason = ERR_SERVER_ERROR
		res.StatusCode = 500
		return res
	}
	if !res.Ok && res.Reason != "" && res.Error == nil {
		res.Error = fmt.Errorf(res.Reason)
	}

	return res
}

func HttpReqJson(method, uri string, body []byte, args map[string]string, headers map[string]string, timeout time.Duration) *OkJson {
	res := &OkJson{Ok: false, Reason: ERR_SERVER_ERROR}
	res_http := HttpReqInternal(method, uri, body, args, headers, timeout)
	res.Error = res_http.Error
	res.RawBody = res_http.RawBody
	res.StatusCode = res_http.StatusCode
	res.Headers = res_http.Headers
	res.ReqDebug = res_http.ReqDebug
	res.Stats = res_http.Stats

	if res_http.StatusCode >= 500 {
		res.Reason = ERR_SERVER_ERROR
		return res
	}

	OkJsonParse(res)

	return res
}

func HttpGetJson(uri string, headers map[string]string, timeout time.Duration) *OkJson {
	return HttpReqJson("GET", uri, nil, nil, headers, timeout)
}

func HttpPostJson(uri string, body []byte, headers map[string]string, timeout time.Duration) *OkJson {
	return HttpReqJson("POST", uri, body, nil, headers, timeout)
}
