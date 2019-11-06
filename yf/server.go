package yf

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jay-wlj/gobaselib/db"
	"github.com/jie123108/glog"
)

const (
	HTTP_GET     = "GET"
	HTTP_POST    = "POST"
	HTTP_OPTIONS = "OPTIONS"
)

type RouterInfo struct {
	Op         string
	Url        string
	Checksign  bool
	Checktoken bool
	Handler    gin.HandlerFunc
}

type Config struct {
	Addr       string
	Debug      bool
	CheckSign  bool
	AppKeys    map[string]string
	AuthServer string
}

type fnRouter func() (prefix string, vs []RouterInfo)

type routers []RouterInfo

func (r *routers) GetIgnoreSignList() (mp map[string]bool) {
	mp = make(map[string]bool)
	for _, v := range *r {
		if !v.Checksign {
			mp[v.Url] = true
		}
	}
	return
}

func (r *routers) GetTokenList() (mp map[string]bool) {
	mp = make(map[string]bool)
	for _, v := range *r {
		if v.Checktoken {
			mp[v.Url] = true
		}
	}
	return
}

type Server struct {
	*gin.Engine
	//mRouter map[string]*RouterInfo

	frouters []fnRouter
}

func NewServer() *Server {
	return &Server{Engine: gin.Default()}
}

func (t *Server) AddRouter(f fnRouter) {
	t.frouters = append(t.frouters, f)
}

func (t *Server) Start(cfg *Config) error {

	t.Use(Cors)
	t.Use(Sign_Check)
	t.Use(Token_Check)
	t.Use(t.handlerwrap) // 处理请求中间件

	var rs routers
	for _, f := range t.frouters {
		prefix, vs := f()
		t.routerRegister(prefix, vs)

		for i := range vs {
			vs[i].Url = prefix + vs[i].Url
		}
		rs = append(rs, vs...)
	}

	SignConfig.Debug = cfg.Debug
	SignConfig.CheckSign = cfg.CheckSign
	SignConfig.AppKeys = cfg.AppKeys
	SignConfig.IgnoreSignList = rs.GetIgnoreSignList()

	TokenConfig.AccountServer = cfg.AuthServer
	TokenConfig.Debug = cfg.Debug
	TokenConfig.NeedTokenList = rs.GetTokenList()

	//t.Use(middleware...)

	server := http.Server{Addr: cfg.Addr, Handler: t.Engine}
	go server.ListenAndServe()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)
	select {
	case <-ch:
		println("sutodwn...")
		timeout := 5 * time.Second
		now := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := server.Shutdown(ctx)
		if err != nil {
			fmt.Println("err:", err)
		}
		fmt.Println("-----exited------", time.Since(now))

		//t.
	}
	return nil
}

func (t *Server) routerRegister(prefix string, vs []RouterInfo) {
	var g *gin.RouterGroup = &t.Engine.RouterGroup

	if prefix != "" {
		g = t.Group(prefix)
	}
	for _, v := range vs {
		g.Handle(v.Op, v.Url, v.Handler)
		// switch v.Op {
		// case common.HTTP_GET:
		// 	g.GET(v.Url, v.Handler)
		// case common.HTTP_POST:
		// 	g.POST(v.Url, v.Handler)
		// }
	}
}

func (t *Server) handlerwrap(c *gin.Context) {

	c.Next() // 执行处理函数

	// 获取事务db
	var sqldb *db.PsqlDB
	if conn, exist := c.Get("sqldao"); exist {
		if db, ok := conn.(*db.PsqlDB); ok {
			sqldb = db
		}
	}

	if sqldb != nil {
		tx := GetRespTx(c)
		if tx {
			err := sqldb.Commit().Error
			if err != nil {
				glog.Errorf("commit err! %v", err)
			}
		} else {
			sqldb.Rollback()
		}
	}

	return
}
