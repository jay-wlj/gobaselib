package yf

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/levigross/grequests"
)

// 异步请求客户端
type AsyncHttp struct {
	host string
	*grequests.Session
}

type OkJson struct {
	Ok   bool        `json:"ok"`
	Err  string      `json:"reason"`
	Data interface{} `json:"data,omitempty"`
}

func NewAsyncHttp(host string, timeout time.Duration) *AsyncHttp {
	return &AsyncHttp{
		host: host,
		Session: grequests.NewSession(&grequests.RequestOptions{
			UserAgent:      "nicefilm/v1.0",
			RequestTimeout: timeout,
		}),
	}
}

func (a *AsyncHttp) Url(path string) string {
	return a.host + path
}

func (a *AsyncHttp) Do(path string, data map[string]interface{}, out interface{}) error {
	url := a.Url(path)
	resp, err := a.Post(url, &grequests.RequestOptions{
		JSON: data,
	})

	if err != nil {
		return err
	}
	defer resp.Close()

	var jso OkJson
	err = resp.JSON(&jso)
	if err != nil {
		return err
	}
	if !jso.Ok {
		return errors.New(jso.Err)
	}

	if out != nil {
		body, _ := json.Marshal(jso.Data)
		return json.Unmarshal(body, out)
	}
	return nil
}

// 添加或者更新命名路由
// name string 路由名, 必填
// url string 请求地址，必填
// method string 请求方法，默认为post
// contentType string 默认为 application/x-www-form-urlencode
// maxRetry int 默认为10， 最大重试次数 3 <= max_retry <= 20
func (a *AsyncHttp) UpsertOption(name, url, method, contentType string, maxRetry int) error {
	data := map[string]interface{}{
		"name":         name,
		"url":          url,
		"method":       method,
		"content_type": contentType,
		"max_retry":    maxRetry,
	}
	return a.Do("/options/upsert", data, nil)
}

// 添加签名路由
// headers 需要指定X-Yf-Appid
func (a *AsyncHttp) UpsertOptionSign(name, url, method, contentType string, maxRetry int, headers map[string]string) error {
	data := map[string]interface{}{
		"name":         name,
		"url":          url,
		"method":       method,
		"content_type": contentType,
		"max_retry":    maxRetry,
		"sign":         true,
		"headers":      headers,
	}
	return a.Do("/options/upsert", data, nil)
}

// 发送消息
// name string 路由名, 必填
// body interface 请求body，选填
// delay int64 延迟时间，选填
// send_time int64 制定发送时间，必须大于当前时间戳或者为0
// headers map 选填
func (a *AsyncHttp) PublishMessage(name string, body interface{}, delay, sendTime int64, headers map[string]string) error {
	data := map[string]interface{}{
		"name":      name,
		"body":      body,
		"delay":     delay,
		"send_time": sendTime,
		"headers":   headers,
	}
	return a.Do("/messages/publish", data, nil)
}

// 立刻发送消息
func (a *AsyncHttp) PublishMessageNow(name string, body interface{}) error {
	return a.PublishMessage(name, body, 0, 0, nil)
}

var (
	DefaultAsyncHttp = NewAsyncHttp("http://127.0.0.1:9401/v1", time.Second*10)
)

func UpsertOption(name, url, method, contentType string, maxRetry int) error {
	return DefaultAsyncHttp.UpsertOption(name, url, method, contentType, maxRetry)
}

func PublishMessage(name string, body interface{}, delay, sendTime int64, headers map[string]string) error {
	return DefaultAsyncHttp.PublishMessage(name, body, delay, sendTime, headers)
}

func PublishMessageNow(name string, body interface{}) error {
	return DefaultAsyncHttp.PublishMessageNow(name, body)
}
