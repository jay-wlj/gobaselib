package yf

import "bytes"

type ApiSign struct {
	AppId  string
	AppKey string
}

// 签名
func (t *ApiSign) Sign(uri string, args, headers map[string][]string, body []byte) (signature, sign string) {
	path, err := GetUriPath(uri)
	if err != nil {
		path = uri
	}

	if body == nil {
		body = EMPTY_BODY
	}
	Sha1Body := Sha1hex(body)
	CanonicalURI := URI_ENCODE(path)
	CanonicalArgs := createCanonicalArgs(args)
	CanonicalHeaders, SignedHeaders := createCanonicalHeaders(headers)

	signbuf := bytes.Buffer{}
	signbuf.WriteString(CanonicalURI)
	signbuf.WriteString("\n")
	signbuf.WriteString(CanonicalArgs)
	signbuf.WriteString("\n")
	signbuf.WriteString(CanonicalHeaders)
	signbuf.WriteString("\n")
	signbuf.WriteString(SignedHeaders)
	signbuf.WriteString("\n")
	signbuf.WriteString(Sha1Body)
	signbuf.WriteString("\n")
	signbuf.WriteString(t.AppKey)

	// SignStr := CanonicalURI + "\n" +
	// 	CanonicalArgs + "\n" +
	// 	CanonicalHeaders + "\n" +
	// 	SignedHeaders + "\n" +
	// 	Sha1Body + "\n" +
	// 	t.AppKey

	signature = signbuf.String()
	sign = Sha1hex([]byte(signature)) // sha加密
	return
}
