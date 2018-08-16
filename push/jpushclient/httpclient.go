package jpushclient

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	CHARSET                    = "UTF-8"
	CONTENT_TYPE_JSON          = "application/json"
	DEFAULT_CONNECTION_TIMEOUT = 20 //seconds
	DEFAULT_SOCKET_TIMEOUT     = 30 // seconds
)

func SendGet(url string, param map[string]string, authcode string) (string, error) {
	req := Get(url)
	req.SetTimeout(DEFAULT_CONNECTION_TIMEOUT*time.Second, DEFAULT_SOCKET_TIMEOUT*time.Second)
	//q.Header("Connection", "Keep-Alive")
	req.Header("Charset", CHARSET)
	req.Header("Authorization", authcode)
	req.Header("Accept", CONTENT_TYPE_JSON)

	for key, value := range param {
		req.Param(key, value)
	}

	fmt.Printf("url:%v req:%v", url, req)

	return req.String()
}

func SendPostString(url, content, authCode string) (string, error) {

	//req := Post(url).Debug(true)
	req := Post(url)
	req.SetTimeout(DEFAULT_CONNECTION_TIMEOUT*time.Second, DEFAULT_SOCKET_TIMEOUT*time.Second)
	req.Header("Connection", "Keep-Alive")
	req.Header("Charset", CHARSET)
	req.Header("Authorization", authCode)
	req.Header("Content-Type", CONTENT_TYPE_JSON)
	req.SetProtocolVersion("HTTP/1.1")
	req.Body(content)

	return req.String()
}

func SendPostBytes(url string, content []byte, authCode string) (string, error) {

	req := Post(url)
	req.SetTimeout(DEFAULT_CONNECTION_TIMEOUT*time.Second, DEFAULT_SOCKET_TIMEOUT*time.Second)
	req.Header("Connection", "Keep-Alive")
	req.Header("Charset", CHARSET)
	req.Header("Authorization", authCode)
	req.Header("Content-Type", CONTENT_TYPE_JSON)
	req.SetProtocolVersion("HTTP/1.1")
	req.Body(content)

	return req.String()
}

func SendPostBytes2(url string, data []byte, authCode string) (string, error) {

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Add("Charset", CHARSET)
	req.Header.Add("Authorization", authCode)
	req.Header.Add("Content-Type", CONTENT_TYPE_JSON)
	resp, err := client.Do(req)

	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}
		return "", err
	}
	if resp == nil {
		return "", nil
	}

	defer resp.Body.Close()
	r, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(r), nil
}
