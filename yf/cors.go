package yf

import (
	"github.com/gin-gonic/gin"
	//"github.com/jie123108/glog"
	"net/http"
)

func Cors(c *gin.Context) {
	method := c.Request.Method
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")	// GET POST方法都需要加此设置
	//放行所有OPTIONS方法
	if method == "OPTIONS" {
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			headerStr := "Access-Control-Allow-Origin, Access-Control-Allow-Headers, Origin, X-Requested-With, Content-Type, Accept, Range, X-Yf-Platform, X-Yf-Appid,X-YF-Rid,X-YF-Sign, X-YF-Version,X-Yf-Token, X-Yf-hash, X-Yf-filesize, X-Yf-chunksize, X-Yf-chunkindex, X-Yf-chunkhash, X-Yf-filename,X-Yf-imgsize, callback"

			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Headers", headerStr)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			// c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			//c.Writer.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
			// c.Header("Access-Control-Max-Age", "172800")
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			c.Writer.Header().Set("content-type", "application/json")
		}
		c.AbortWithStatus(http.StatusOK)
		return
	}

	c.Next()
}