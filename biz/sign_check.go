package biz

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"strings"
	base "gobaselib"
	// . "gobaselib/common"
)

type ApiSignConfig struct {
	Debug          bool
	CheckSign      bool
	DebugSignKey   string
	AppKeys        map[string]string
	IgnoreSignList map[string]bool
}

var SignConfig ApiSignConfig = ApiSignConfig{false, true, "eba0cb9dc8cffd641be5d01969674a30", make(map[string]string), make(map[string]bool)}

func (this *ApiSignConfig) GetSignKey(appid string) string {
	return this.AppKeys[appid]
}

func ApiSignCheck(c *gin.Context, body []byte) bool {
	c.Request.ParseForm()
	uri := c.Request.RequestURI
	args := c.Request.Form
	headers := c.Request.Header

	app_ids := headers["X-Mt-Appid"]
	if len(app_ids) != 1 {
		glog.Errorf("find %d AppId value..", len(app_ids))
		c.JSON(401, gin.H{"ok": false, "reason": ERR_ARGS_INVALID})
		c.Abort()
		return false
	}
	app_id := app_ids[0]
	c.Set("app_id", app_id)

	app_key := SignConfig.GetSignKey(app_id)
	if app_key == "" {
		glog.Errorf("Unknow appid [%s]", app_id)
		c.JSON(401, gin.H{"ok": false, "reason": ERR_ARGS_INVALID})
		c.Abort()
		return false
	}

	req_signs := headers["X-Mt-Sign"]
	if len(req_signs) != 1 {
		glog.Errorf("find %d Sign value..", len(req_signs))
		c.JSON(401, gin.H{"ok": false, "reason": ERR_SIGN_ERROR})
		c.Abort()
		return false
	}
	req_sign := req_signs[0]
	// glog.Errorf("req_sign: %s", req_sign)
	// 测试工具使用。
	if req_sign == SignConfig.DebugSignKey && SignConfig.Debug {
		return true
	}

	signature, SignStr := Sign(uri, args, headers, body, app_key)
	if signature != req_sign {
		glog.Errorf("req_sign: [%s] != calc_sign: [%s] \nSignStr [[%s]]", req_sign,
			signature, SignStr)
		glog.Infof("req body len: %d", len(body))
		if SignConfig.Debug && len(body) < 100 {
			glog.Infof("body: [[%v]]", string(body))
		}
		c.JSON(401, gin.H{"ok": false, "reason": ERR_SIGN_ERROR, "SignStr": SignStr})
		c.Abort()
		return false
	}
	return true
}

func SignCheck(c *gin.Context) {
	uri := base.GetUri(c)

	body := []byte("")
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		body, _ = ioutil.ReadAll(c.Request.Body)
		c.Set("body", body)
	}

	if SignConfig.CheckSign && !SignConfig.IgnoreSignList[uri] {
		ApiSignCheck(c, body)
	}

	c.Next()

	//context.Clear(c.Request)
}
