package mt

import (
	"gobaselib/log"
	"io/ioutil"

	"github.com/gin-gonic/gin"

	//"strings"
	base "github.com/jay-wlj/gobaselib"
)

type ApiSignConfig struct {
	Debug          bool
	CheckSign      bool
	DebugSignKey   string
	AppKeys        map[string]string
	IgnoreSignList map[string]bool
}

var SignConfig ApiSignConfig = ApiSignConfig{false, true, "62361670a0b60c852fcc1e69189c233e", make(map[string]string), make(map[string]bool)}

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
		log.Errorf("find %d AppId value..", len(app_ids))
		c.JSON(401, gin.H{"ok": false, "reason": ERR_ARGS_INVALID})
		c.Abort()
		return false
	}
	app_id := app_ids[0]
	c.Set("app_id", app_id)

	app_key := SignConfig.GetSignKey(app_id)
	if app_key == "" {
		log.Errorf("Unknow appid [%s]", app_id)
		c.JSON(401, gin.H{"ok": false, "reason": ERR_ARGS_INVALID})
		c.Abort()
		return false
	}

	req_signs := headers["X-Mt-Sign"]
	if len(req_signs) != 1 {
		log.Errorf("find %d Sign value..", len(req_signs))
		c.JSON(401, gin.H{"ok": false, "reason": ERR_SIGN_ERROR})
		c.Abort()
		return false
	}
	req_sign := req_signs[0]
	// log.Errorf("req_sign: %s", req_sign)
	// 测试工具使用。
	if req_sign == SignConfig.DebugSignKey && SignConfig.Debug {
		return true
	}

	signature, SignStr := Sign(uri, args, headers, body, app_key)
	if signature != req_sign {
		log.Errorf("req_sign: [%s] != calc_sign: [%s] \nSignStr [[%s]]", req_sign,
			signature, SignStr)
		log.Infof("req body len: %d", len(body))
		if SignConfig.Debug && len(body) < 100 {
			log.Infof("body: [[%v]]", string(body))
		}
		c.JSON(401, gin.H{"ok": false, "reason": ERR_SIGN_ERROR, "SignStr": SignStr})
		c.Abort()
		return false
	}
	return true
}

func Sign_Check(c *gin.Context) {
	uri := base.GetUri(c)

	body := []byte("")
	if c.Request.Method == "POST" || c.Request.Method == "PUT" {
		body, _ = ioutil.ReadAll(c.Request.Body)
		c.Set("viewbody", body)
	}

	if SignConfig.CheckSign && !SignConfig.IgnoreSignList[uri] {
		ApiSignCheck(c, body)
	}

	c.Next()

	//context.Clear(c.Request)
}
