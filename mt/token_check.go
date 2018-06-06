package mt

import (
	// "fmt"
	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	//"strings"
	base "gobaselib"
	"time"
)

// http://www.gorillatoolkit.org/pkg/context

type ApiTokenConfig struct {
	Debug          bool
	CheckSign      bool
	AccountServer  string
	AccountTimeout time.Duration
	NeedTokenList  map[string]bool
}

var TokenConfig ApiTokenConfig = ApiTokenConfig{false, true, "http://127.0.0.1:810", 5 * time.Second, make(map[string]bool)}

func token_check(c *gin.Context) bool {
	c.Request.ParseForm()
	// uri := c.Request.RequestURI
	// args := c.Request.Form
	headers := c.Request.Header
	// app_key := common.Config.AppKey

	req_tokens := headers["X-Mt-Token"]
	if len(req_tokens) != 1 {
		glog.Errorf("find %d Token value..", len(req_tokens))
		c.JSON(401, gin.H{"ok": false, "reason": ERR_TOKEN_INVALID})
		c.Abort()
		return false
	}

	token := req_tokens[0]
	glog.Infof("req_token: %s", token)
	// Check Token from account
	user_id, user_type, _, err := TokenCheck(TokenConfig.AccountServer, token, TokenConfig.AccountTimeout)
	if err != nil {
		glog.Errorf("CheckToken(%s) failed! timeout:%v err: %v", token, TokenConfig.AccountTimeout, err)
		c.JSON(401, gin.H{"ok": false, "reason": ERR_TOKEN_INVALID})
		c.Abort()
		return false
	}
	c.Set("user_id", user_id)
	c.Set("user_type", user_type)
	return true
}

func Token_Check(c *gin.Context) {
	uri := base.GetUri(c)

	if !TokenConfig.NeedTokenList[uri] || c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	app_id, isexist := c.Get("app_id")
	if isexist && app_id == nil {
		headers := c.Request.Header
		app_ids := headers["X-Mt-Appid"]
		if len(app_ids) != 1 {
			glog.Errorf("find %d AppId value..", len(app_ids))
			c.JSON(401, gin.H{"ok": false, "reason": ERR_ARGS_INVALID})
			c.Abort()
			return
		}
		app_id := app_ids[0]
		c.Set("app_id", app_id)
	}

	//if TokenConfig.CheckSign {
	token_check(c)
	//}

	c.Next()
	//context.Clear(c.Request)
	//c.Clear()
}
