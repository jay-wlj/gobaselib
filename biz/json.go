package biz

import (
	"github.com/gin-gonic/gin"
)

// func JsonOk(data interface{}) gin.H {
// 	return gin.H{"ok": true, "reason": "", "data": data}
// }

// func JsonFail(reason string) gin.H {
// 	return gin.H{"ok": false, "reason": reason}
// }

func JsonOk(c *gin.Context, data interface{}) gin.H {
	response := gin.H{"ok": true, "reason": "", "data": data}
	JSON(c, 200, response)
	return response
}

func JsonFail(c *gin.Context, reason string) gin.H {
	response := gin.H{"ok": false, "reason": reason}
	if reason == ERR_SERVER_ERROR {
		JSON(c, 500, response)
	} else {
		JSON(c, 200, response)
	}
	return response
}

func JSON(c *gin.Context, code int, obj interface{}) {
	c.Set("resp_code", code)
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
