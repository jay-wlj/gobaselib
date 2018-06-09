package gobaselib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"github.com/json-iterator/go"
	//"unsafe"
)

func GetUri(c *gin.Context) string {
	uri := c.Request.RequestURI

	pos := strings.Index(uri, "?")
	if pos >= 0 && pos < len(uri) {
		uri = uri[0:pos]
	}
	return uri
}

func GetPostJsonData(c *gin.Context) ([]byte, error) {
	body, exists := c.Get("body")
	buf := []byte{}
	var err error
	if !exists {
		raw_body := c.Request.Body
		buf, err = ioutil.ReadAll(raw_body)
	} else {
		buf = body.([]byte)
	}
	return buf, err
}

func GetQueryJsonObject(c *gin.Context, query interface{}) (err error) {
	buf, err := GetPostJsonData(c)

	if err != nil {
		glog.Errorf("Invalid body[%v] err: %v", string(buf), err)
		return
	}

	glog.Infof("uri:%v buf:%v", GetUri(c), string(buf))

	err = json.Unmarshal(buf, query)
	if err != nil {
		PostForm := c.Request.PostForm
		Form := c.Request.Form
		glog.Errorf("Invalid body[%v] PostForm[%v] Form[%v], err: %v", string(buf), PostForm, Form, err)
		return
	}

	return
}

func QueryInt(c *gin.Context, key string) (value int, err error) {
	strvalue := c.Query(key)
	if strvalue == "" {
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.Atoi(strvalue)
	return
}

func QueryIntDef(c *gin.Context, key string, def int) (value int) {
	strvalue := c.Query(key)
	if strvalue == "" {
		value = def
		return
	}
	var err error
	value, err = strconv.Atoi(strvalue)
	if err != nil {
		glog.Errorf("Invalid int value(%s) err: %v", strvalue, err)
		value = def
	}

	return
}

// func CheckQueryIntDefaultField(c *gin.Context, key string, def int) (value int) {
// 	strvalue := c.Query(key)
// 	if 0 == len(strvalue) {
// 		value = def
// 		return
// 	}
// 	var err error
// 	value, err = strconv.Atoi(strvalue)
// 	if err != nil {
// 		glog.Errorf("invalid int value(%s) ", strvalue)
// 		value = def
// 	}
// 	return
// }

// func CheckQueryInt64Field(c *gin.Context, key string) (value int64, err error) {
// 	strvalue := c.Query(key)
// 	if 0 == len(strvalue) {
// 		value = 0
// 		err = fmt.Errorf("ERR_ARGS_MISSING")
// 		return
// 	}
// 	value, err = strconv.ParseInt(strvalue, 10, 64)
// 	return
// }

// func CheckQueryBoolField(c *gin.Context, key string) (value bool, err error) {
// 	strvalue := c.Query(key)
// 	if 0 == len(strvalue) {
// 		value = false
// 		err = fmt.Errorf("ERR_ARGS_MISSING")
// 		return
// 	}
// 	value, err = strconv.ParseBool(strvalue)
// 	return
// }

// //检测必须的字段是否为空，若都不为空则返回true,反之为false
// func CheckNilField(info interface{}, fields []string) (ret bool) {
// 	defer func() {
// 		if err := recover(); err != nil {
// 			glog.Errorf("%v", err)
// 			ret = true
// 		}
// 	}()
// 	v := reflect.ValueOf(info)
// 	if v.Kind() == reflect.Ptr {
// 		v = v.Elem()
// 	}
// 	for _, k := range fields {
// 		one := v.FieldByName(k)
// 		if !one.IsValid() || one.Interface() == reflect.Zero(one.Type()).Interface() {
// 			glog.Errorf("field %s is nil", k)
// 			return false
// 		}
// 	}
// 	return true
// }
