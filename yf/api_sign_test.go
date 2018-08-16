package yf

import (
	"fmt"
	"github.com/jie123108/glog"
	"strings"
	"testing"
)

func parse_headers(str_headers string) map[string][]string {
	lines := strings.Split(str_headers, "\n")
	headers := make(map[string][]string)
	for _, head := range lines {
		if head == "" {
			continue
		}
		idx := strings.Index(head, ":")
		if idx == -1 {
			glog.Errorf("Invalid Head: %s", head)
		} else {
			key := strings.ToLower(strings.TrimSpace(head[0:idx]))
			value := strings.TrimSpace(head[idx+1:])
			headers[key] = append(headers[key], value)
		}
	}
	return headers
}

func parse_args(str_args string) map[string][]string {
	args := make(map[string][]string)
	as := strings.Split(str_args, "&")
	for _, arg := range as {
		arr := strings.Split(arg, "=")
		key, value := "", ""
		if len(arr) == 2 {
			key = arr[0]
			value = arr[1]
		} else if len(arr) == 1 {
			key = arr[0]
		} else {
			glog.Errorf("---- invalid arg: %s", arg)
		}
		args[key] = append(args[key], value)
	}
	return args
}

var uri, body, app_key, str_args string
var args map[string][]string
var headers map[string][]string
var signature, SignStr string

func init() {
	// method, uri, args, headers, body_encrypt
	app_key = "16317d117c6eceb8b1b0ebb40e506617"
	uri = "/path/test/~-_/99@/中文.doc"
	str_args = "dest=mongo&DEST=MongoEx&aBo=d9&aBo=Ads&name&aBo=a09&aBo=030"
	// str_args := "dest=mongo&DEST=MongoEx&aBo=d9&aBo=Ads"
	body := "This is the body"
	str_headers := `Host: www.yf.com
Content-Type: application/text
Content-Length: 16
range: 0-1000
date: Fri, 18 Dec 2015 06:17:47 GMT
X-YF-Token: test-token
X-YF-AppId: test
X-YF-rid: 001
X-YF-FOO: Dest
X-YF-FOo: Ads
X-YF-Foo: Abort
X-YF-foo: 099
`

	headers := parse_headers(str_headers)
	args := parse_args(str_args)
	signature, SignStr = Sign(uri, args, headers, []byte(body), app_key)
	fmt.Println("------------- 示例 ---------------")
	fmt.Println("app_key:", app_key)
	fmt.Println("uri:", uri)
	fmt.Println("args:", str_args)
	fmt.Println("headers: ", str_headers)
	fmt.Println("body:", body)
	fmt.Println("------------- 结果 ----------------")
	fmt.Println("签名值: " + signature)
	fmt.Println("SignStr: [[" + SignStr + "]]")
}

func Test_Sign(t *testing.T) {
	if signature == "719d7d49eebc533d3480d25685199d38fdc430ef" {
		t.Log("OK")
	} else {
		t.Error("signature failed!")
	}
}

/**

func main() {

	url := "/path/test/~-_/99@/中文.doc"
	fmt.Println(URI_ENCODE(url))
	// fmt.Println(uri_encode(url))
	fmt.Println(URI_ENCODE("/path/test/~-_/99%40/%E4%B8%AD%E6%96%87.doc"))

	args := make(map[string][]string)
	args["test"] = []string{"aaa"}
	args["tEaa"] = []string{"abc", "def"}
	args["aaa"] = []string{"bbb"}
	args["aab"] = []string{"ccc", "ddd", "cbc"}

	fmt.Println(createCanonicalArgs(args))
}

**/
