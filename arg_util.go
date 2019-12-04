package base

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type ArgChecker struct {
	ErrorStr string
	// 是否检查所有
	all bool
}

func NewChecker(all ...bool) *ArgChecker {
	var ckAll bool = true
	if len(all) > 0 {
		ckAll = all[0]
	}
	return &ArgChecker{
		all: ckAll,
	}
}

func (ck *ArgChecker) OK() bool {
	return len(ck.ErrorStr) == 0
}

func (ck *ArgChecker) Error() string {
	return ck.ErrorStr
}

func (ck *ArgChecker) GetInt(c *gin.Context, key string) int {
	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		return 0
	}
	ret, err := strconv.Atoi(data)
	if err != nil { //非整型
		return 0
	}
	return ret
}

func (ck *ArgChecker) Check(c *gin.Context, key string) string {
	if !ck.all && len(ck.ErrorStr) != 0 {
		return ""
	}
	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		ck.ErrorStr = ck.ErrorStr + " " + key + " missing"
		return ""
	}
	if data == "" {
		//参数为空
		ck.ErrorStr = ck.ErrorStr + " " + key + " empty"
		return ""
	}
	return data
}

func (ck *ArgChecker) CheckInt(c *gin.Context, key string) int {
	return ck.CheckIntDef(c, key, 0)
}

func (ck *ArgChecker) CheckIntDef(c *gin.Context, key string, def int) int {
	if !ck.all && len(ck.ErrorStr) != 0 {
		return 0
	}

	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		return def
	}
	ret, err := strconv.Atoi(data)
	if err != nil { //非整型
		return def
	}
	return ret
}

func (ck *ArgChecker) CheckInt64(c *gin.Context, key string) int64 {
	if !ck.all && len(ck.ErrorStr) != 0 {
		return 0
	}

	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		ck.ErrorStr = ck.ErrorStr + " " + key + " missing"
		return 0
	}
	ret, err := strconv.ParseInt(data, 10, 64)
	if err != nil { //非整型
		ck.ErrorStr = ck.ErrorStr + " " + key + err.Error()
		return 0
	}
	return ret
}

func (ck *ArgChecker) CheckIntList(c *gin.Context, key string, sep string) []int {
	if !ck.all && len(ck.ErrorStr) != 0 {
		return nil
	}

	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		ck.ErrorStr = ck.ErrorStr + " " + key + " missing"
		return nil
	}
	ids := strings.Split(data, ",")
	var ret = make([]int, len(ids))
	for idx, it := range ids {
		id, err := strconv.Atoi(it)
		if err != nil { //非整型
			ck.ErrorStr = ck.ErrorStr + " " + key + err.Error()
			return nil
		}
		ret[idx] = id
	}
	return ret
}

func (ck *ArgChecker) CheckInt64List(c *gin.Context, key string, sep string) []int64 {
	if !ck.all && len(ck.ErrorStr) != 0 {
		return nil
	}

	data, ok := c.GetQuery(key)
	if !ok {
		//找不到参数
		ck.ErrorStr = ck.ErrorStr + " " + key + " missing"
		return nil
	}
	ids := strings.Split(data, ",")
	var ret = make([]int64, len(ids))
	for idx, it := range ids {
		id, err := strconv.ParseInt(it, 10, 64)
		if err != nil { //非整型
			ck.ErrorStr = ck.ErrorStr + " " + key + err.Error()
			return nil
		}
		ret[idx] = id
	}
	return ret
}
