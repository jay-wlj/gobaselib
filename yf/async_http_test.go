package yf

import (
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func expect(t *testing.T, a interface{}, b interface{}) {
	if a != b {
		t.Errorf("Expected %v (type %v) - Got %v (type %v)", b, reflect.TypeOf(b), a, reflect.TypeOf(a))
	}
}

func Test_AsyncHTTP(t *testing.T) {
	name := "test.test.test"
	err := UpsertOption(name, "http://127.0.0.1:61211/test", "post", "", 10)
	expect(t, err, nil)

	err = PublishMessageNow(name, "key=value&a=1")
	expect(t, err, nil)
}

func Test_Wait(t *testing.T) {
	var sem = make(chan struct{})
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.POST("/test", func(c *gin.Context) {
		var out struct {
			Key string `form:"key"`
			A   int    `form:"a"`
		}
		err := c.Bind(&out)
		expect(t, err, nil)
		expect(t, out.Key, "value")
		expect(t, out.A, 1)
		close(sem)
	})
	go r.Run(":61211")
	select {
	case <-time.After(time.Second * 3):
		t.Error("receive message timeout")
	case <-sem:
	}
}
