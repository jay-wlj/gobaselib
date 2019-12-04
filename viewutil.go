package base

import (
	//"encoding/json"
	"bytes"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"

	jsoniter "github.com/json-iterator/go"
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

func CheckStringToInt64(strvalue string) (value int64, err error) {
	if 0 == len(strvalue) {
		value = 0
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.ParseInt(strvalue, 10, 64)
	return
}

func CheckStringToInt(strvalue string) (value int, err error) {
	if 0 == len(strvalue) {
		value = 0
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.Atoi(strvalue)
	return
}

func CheckStringToFloat64(strvalue string) (value float64, err error) {
	if 0 == len(strvalue) {
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.ParseFloat(strvalue, 64)
	return
}

func CheckQueryStringField(c *gin.Context, key string) (value string, err error) {
	strvalue := c.Query(key)
	return strvalue, nil
}

func CheckQueryIntField(c *gin.Context, key string) (value int, err error) {
	strvalue := c.Query(key)
	return CheckStringToInt(strvalue)
}

func CheckQueryIntDefaultField(c *gin.Context, key string, def int) (value int, err error) {
	strvalue := c.Query(key)
	if 0 == len(strvalue) {
		value = def
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.Atoi(strvalue)
	return
}

func CheckQueryInt64Field(c *gin.Context, key string) (value int64, err error) {
	strvalue := c.Query(key)
	return CheckStringToInt64(strvalue)
}

func CheckQueryInt64DefaultField(c *gin.Context, key string, def int64) (value int64, err error) {
	strvalue := c.Query(key)
	if 0 == len(strvalue) {
		value = def
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.ParseInt(strvalue, 10, 64)
	return
}

func CheckQueryBoolField(c *gin.Context, key string) (value bool, err error) {
	strvalue := c.Query(key)
	if 0 == len(strvalue) {
		value = false
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.ParseBool(strvalue)
	return
}

func CheckQueryFloat64Field(c *gin.Context, key string) (value float64, err error) {
	strvalue := c.Query(key)
	if 0 == len(strvalue) {
		err = fmt.Errorf("ERR_ARGS_MISSING")
		return
	}
	value, err = strconv.ParseFloat(strvalue, 64)
	return
}

func GetPostJsonData(c *gin.Context) ([]byte, error) {
	body, exists := c.Get("viewbody")
	buf := []byte{}
	var err error
	if !exists {
		if c.Request.Body != nil {
			buf, _ = ioutil.ReadAll(c.Request.Body)
			c.Request.Body = ioutil.NopCloser(bytes.NewReader(buf))
		}
		c.Set("viewbody", buf) // 注 需要保存body
		//glog.Info("GetPostJsonData buf:", string(buf))
	} else {
		buf = body.([]byte)
	}
	return buf, err
}

func CheckQueryJsonField(c *gin.Context, stu interface{}) error {
	buf, err := GetPostJsonData(c)

	if err == nil {
		glog.Infof("uri:%v buf:%v", GetUri(c), string(buf))
		//err = json.Unmarshal(buf, stu)
		err = jsoniter.Unmarshal(buf, stu)
		if err != nil {
			post_form := c.Request.PostForm
			form := c.Request.Form
			glog.Errorf("1. try Invalid body[%v] postform[%v] form[%v], err: %v", string(buf), post_form, form, err)
		}
	} else {
		glog.Errorf("2. try  Invalid body[%v] err: %v", string(buf), err)
	}

	return err
}

//检测必须的字段是否为空，若都不为空则返回true,反之为false
func CheckNilField(info interface{}, fields []string) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			glog.Errorf("%v", err)
			ret = true
		}
	}()
	v := reflect.ValueOf(info)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for _, k := range fields {
		one := v.FieldByName(k)
		if !one.IsValid() || one.Interface() == reflect.Zero(one.Type()).Interface() {
			glog.Errorf("field %s is nil", k)
			return false
		}
	}
	return true
}
