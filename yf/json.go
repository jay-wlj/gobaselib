package yf

import (
	"github.com/gin-gonic/gin"
)

func JSON_Ok(c *gin.Context, data interface{}) gin.H {
	response := gin.H{"ok": true, "reason": "", "data": data}
	JSON(c, 200, true, response)
	return response
}

func JSON_Fail(c *gin.Context, reason string, err_msg ...string) gin.H {
	response := gin.H{"ok": false, "reason": reason}
	if len(err_msg) > 0 {
		response["err_msg"] = err_msg
	}
	if reason == ERR_SERVER_ERROR {
		JSON(c, 500, false, response)
	} else {
		JSON(c, 200, false, response)
	}
	return response
}

func JSON_FailCode(c *gin.Context, reason string, code int32) gin.H {
	response := gin.H{"ok": false, "reason": reason, "code": code}
	JSON(c, 200, false, response)
	return response
}

func JSON_FailEx(c *gin.Context, reason string, data interface{}) gin.H {
	response := gin.H{"ok": false, "reason": reason, "data": data}
	if reason == ERR_SERVER_ERROR {
		JSON(c, 500, false, response)
	} else {
		JSON(c, 200, false, response)
	}
	return response
}

func JSON(c *gin.Context, code int, txcommit bool, obj interface{}) {
	c.Set("resp_code", code)
	c.Set("resp_tx", txcommit)
	c.JSON(code, obj)
}

func GetRespCode(c *gin.Context) (code int) {
	value, bexist := c.Get("resp_code")
	if bexist {
		code = value.(int)
	} else {
		code = 0
	}
	return code
}
func GetRespTx(c *gin.Context) (tx bool) {
	value, bexist := c.Get("resp_tx")
	if bexist {
		tx = value.(bool)
	} else {
		tx = false
	}
	return
}
