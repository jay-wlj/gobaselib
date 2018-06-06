package mt

import (
	"fmt"
	base "gobaselib"
	"net/url"
	"time"
)

func GetContentType(filename string) string {
	return get_content_type(filename)
}

func ParseArgs(uri string) (args map[string]string) {
	u, err := url.Parse(uri)
	if err != nil {
		return
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err == nil {
		args = make(map[string]string)
		for k, varr := range values {
			args[k] = varr[0] // 有多个参数的, 只取第一个.所以请不要传入多个相同的参数, 会导致签名错误.
		}
		return
	}
	return
}

func YgHttpPost(uri string, body []byte, headers map[string]string, timeout time.Duration, app_key string) *base.OkJson {
	signature, SignStr := "", ""
	if app_key != "" {
		args := ParseArgs(uri)
		signature, SignStr = Sign2(uri, args, headers, body, app_key)
		headers["X-Mt-SIGN"] = signature
	}

	res := base.HttpPostJson(uri, body, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}

func YfHttpGet(uri string, headers map[string]string, timeout time.Duration, app_key string) *base.OkJson {
	signature, SignStr := "", ""
	if app_key != "" {
		args := ParseArgs(uri)
		signature, SignStr = Sign2(uri, args, headers, nil, app_key)
		headers["X-Mt-SIGN"] = signature
	}

	res := base.HttpGetJson(uri, headers, timeout)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}

func CachedNfHttpGet(client *base.RedisHttpClient, exptime time.Duration, uri string, headers map[string]string, timeout time.Duration, app_key string) *base.OkJson {
	signature, SignStr := "", ""
	if app_key != "" {
		args := ParseArgs(uri)
		signature, SignStr = Sign2(uri, args, headers, nil, app_key)
		headers["X-Mt-SIGN"] = signature
	}

	res := client.HttpGetJson(uri, headers, timeout, exptime)

	if res.StatusCode == 401 {
		fmt.Printf("signature [%s] SigStr [[\n%s\n]]", signature, SignStr)
	}
	return res
}
