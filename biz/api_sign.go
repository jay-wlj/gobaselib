package biz

import (
	"bytes"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var pattern *regexp.Regexp = regexp.MustCompile("\\%[0-9A-Fa-f]{2}")

func IsEncoded(str string) bool {
	find := pattern.FindString(str)
	return find != ""
}

func uri_encode_internal(arg string, encodeSlash bool) string {
	if arg == "" || IsEncoded(arg) {
		return arg
	}

	chars := bytes.NewBuffer([]byte{})

	barg := []byte(arg)
	for _, ch := range barg {
		if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-' || ch == '~' || ch == '.' {
			chars.WriteByte(ch)
		} else if ch == '/' {
			if encodeSlash {
				chars.WriteString("%2F")
			} else {
				chars.WriteByte(ch)
			}
		} else {
			chars.WriteString(fmt.Sprintf("%%%02X", ch))
		}
	}

	return chars.String()
}

func URI_ENCODE(uri string) string {
	return uri_encode_internal(uri, false)
}

func uri_encode(uri string) string {
	return uri_encode_internal(uri, true)
}

func createCanonicalArgs(args map[string][]string) string {
	if args == nil {
		return ""
	}
	var keys []string

	for k, _ := range args {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	key_values := []string{}

	for _, key := range keys {
		value := args[key]
		if len(value) == 1 {
			key_values = append(key_values, uri_encode(key)+"="+uri_encode(value[0]))
			// key_values.WriteString()
		} else { // is array
			sort.Strings(value)
			for _, value_sub := range value {
				// key_values.WriteString(uri_encode(key) + "=" + uri_encode(value_sub))
				key_values = append(key_values, uri_encode(key)+"="+uri_encode(value_sub))
			}
		}
	}

	return strings.Join(key_values, "&")
}

// map[string][]string
func createCanonicalHeaders(headers map[string][]string) (string, string) {

	headers_lower := make(map[string]string)
	signed_headers := []string{}

	for k, v := range headers {
		k = strings.ToLower(k)
		// TODO: value 是数组的情况。
		if k != "x-mt-sign" && strings.HasPrefix(k, "x-mt-") {
			signed_headers = append(signed_headers, k)
			sort.Strings(v)
			headers_lower[k] = strings.Join(v, ",")
		}
	}

	header_values := bytes.NewBuffer([]byte{})
	sort.Strings(signed_headers)
	for i, k := range signed_headers {
		if i != 0 {
			header_values.WriteString("\n")
		}
		header_values.WriteString(k + ":" + strings.TrimSpace(headers_lower[k]))
	}
	return header_values.String(), strings.Join(signed_headers, ";")
}

func createSignStr(uri string, args map[string][]string, headers map[string][]string, Sha1Body, app_key string) string {

	CanonicalURI := URI_ENCODE(uri)
	CanonicalArgs := createCanonicalArgs(args)
	CanonicalHeaders, SignedHeaders := createCanonicalHeaders(headers)

	SignStr := CanonicalURI + "\n" +
		CanonicalArgs + "\n" +
		CanonicalHeaders + "\n" +
		SignedHeaders + "\n" +
		Sha1Body + "\n" +
		app_key

	return SignStr
}

func GetUriPath(uri string) (string, error) {
	myurl, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	return myurl.Path, err
}

var EMPTY_BODY []byte = []byte("")

func Sign(uri string, args map[string][]string, headers map[string][]string, body []byte, app_key string) (string, string) {
	path, err := GetUriPath(uri)
	if err != nil {
		path = uri
	}

	if body == nil {
		body = EMPTY_BODY
	}
	Sha1Body := Sha1hex(body)
	SignStr := createSignStr(path, args, headers, Sha1Body, app_key)
	btSignStr := []byte(SignStr)
	signature := Sha1hex(btSignStr)

	return signature, SignStr
}

func Sign2(uri string, args map[string]string, headers map[string]string, body []byte, app_key string) (string, string) {
	myargs := make(map[string][]string)
	for k, v := range args {
		value := make([]string, 1)
		value[0] = v
		myargs[k] = value
	}

	myheaders := make(map[string][]string)
	for k, v := range headers {
		value := make([]string, 1)
		value[0] = v
		myheaders[k] = value
	}
	return Sign(uri, myargs, myheaders, body, app_key)
}

func Sha1(body []byte) string {
	sha1 := Sha1hex(body)
	return string(sha1[:])
}
