package base

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/log"
	"testing"

	"github.com/jay-wlj/gobaselib/cache"
)

type Conf struct {
	Imports []string                  `yaml:"imports"`
	Redis   map[string]cache.RedisCfg `yaml:"redis"`
	AppKeys map[string]string
}

func (t Conf) GetImports() (files []string) {
	return t.Imports
}

func TestLoadConf(t *testing.T) {
	var v Conf

	if err := LoadConf("E:\\Project\\go\\src\\yunbay\\ybgoods\\conf\\config.yml", &v); err != nil {
		log.Error("配置文件读取错误! err=", err)
		return
	}

	fmt.Println("conf=", v)
	return
}
