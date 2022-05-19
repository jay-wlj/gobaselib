package base

import (
	"fmt"
	"github.com/jay-wlj/gobaselib/log"
	"io/ioutil"
	"path/filepath"

	jsoniter "github.com/json-iterator/go"

	yaml "gopkg.in/yaml.v2"
)

type IConf interface {
	GetImports() (files []string) // 获取导入的文件列表并依次加载
}

// 解析配置文件到变量中
func LoadConf(file string, pObj interface{}) (err error) {
	// 先转为绝对路径
	if !filepath.IsAbs(file) {
		file = GetAppPath() + file // 获取配置文件绝对路径
	}
	var data []byte
	if data, err = ioutil.ReadFile(file); err != nil {
		log.Error("配置文件读取错误! err=", err)
		return
	}

	//把yaml文件解析成struct类型
	if err = yaml.Unmarshal(data, pObj); err != nil {
		log.Error("配置信息读取失败!", err)
		return err
	}

	// 优先加载父节点配置
	if iObj, ok := pObj.(IConf); ok {
		imports := iObj.GetImports()
		if len(imports) > 0 {
			for _, v := range imports {
				if !filepath.IsAbs(v) {
					dir, _ := filepath.Split(file) // 取当前配置文件的目录
					v = dir + v                    // 相对于配置文件的路径
				}
				if data, err = ioutil.ReadFile(v); err != nil {
					log.Error("LoadConfig fail! path=", v, " err=", err)
					return err
				}
				if err = yaml.Unmarshal(data, pObj); err != nil {
					log.Error("LoadConfig fail! path=", v, " err=", err)
					return err
				}
			}

			// 覆盖父亲节点的一些配置
			if data, err = ioutil.ReadFile(file); err != nil {
				log.Error("配置文件读取错误! err=", err)
				return
			}
			err = yaml.Unmarshal(data, pObj)
			if err != nil {
				log.Error("配置信息读取失败!", err)
				return err
			}
		}
	}

	if data, err = jsoniter.MarshalIndent(pObj, "", "  "); err != nil {
		log.Error("配置信息读取失败!", err)
		return
	}
	fmt.Println("--------load conf-------\r\n", string(data))
	return err
}
