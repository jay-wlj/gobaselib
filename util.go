package base

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Maps map[string]interface{}

const FLOAT_MIN = 0.0000001
const FLOAT_MIN_PRECISION = 8

func StringToInt(str string) (value int, err error) {
	value, err = strconv.Atoi(str)
	return value, err
}

func StringToInt64(str string) (value int64, err error) {
	value, err = strconv.ParseInt(str, 10, 64)
	return
}

func StringToFloat64(str string) (value float64, err error) {
	value, err = strconv.ParseFloat(str, 64)
	return
}

func IntToString(value int) (strvalue string) {
	strvalue = strconv.Itoa(value)
	return
}

func Int64ToString(value int64) (strvalue string) {
	strvalue = strconv.FormatInt(value, 10)
	return
}

func Uint64ToString(value uint64) (strvalue string) {
	strvalue = strconv.FormatUint(value, 10)
	return
}

func Float64ToString(value float64) (strvalue string) {
	strvalue = strconv.FormatFloat(value, 'f', -1, 64)
	return
}

func IntToInt64(val int) (value int64, err error) {
	strval := IntToString(val)
	value, err = StringToInt64(strval)
	return
}

func Int64ToInt(val int64) (value int, err error) {
	strval := Int64ToString(val)
	value, err = strconv.Atoi(strval)
	return
}
func IntSliceToString(values []int, splite string) (strvalue string) {
	bfirst := true
	for _, value := range values {
		if !bfirst {
			strvalue += splite
		} else {
			bfirst = false
		}
		strvalue += IntToString(value)
	}
	return
}

func StringSliceToString(values []string, splite string) (strvalue string) {
	bfirst := true
	for _, value := range values {
		if !bfirst {
			strvalue += splite
		} else {
			bfirst = false
		}
		strvalue += value
	}
	return
}

func Int64SliceToString(values []int64, splite string) (strvalue string) {
	bfirst := true
	for _, value := range values {
		if !bfirst {
			strvalue += splite
		} else {
			bfirst = false
		}
		strvalue += Int64ToString(value)
	}
	return
}

func Uint64SliceToString(values []uint64, splite string) (strvalue string) {
	bfirst := true
	for _, value := range values {
		if !bfirst {
			strvalue += splite
		} else {
			bfirst = false
		}
		strvalue += Uint64ToString(value)
	}
	return
}
func StringToIntSlice(str string, splite string) (ivalues []int) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	strarr := strings.Split(str, splite)
	for _, strvalue := range strarr {
		ivalue, _ := StringToInt(strvalue)
		ivalues = append(ivalues, ivalue)
	}
	return
}

func StringToInt64Slice(str string, splite string) (ivalues []int64) {
	str = strings.TrimSpace(str)
	if str == "" {
		return nil
	}
	strarr := strings.Split(str, splite)
	for _, strvalue := range strarr {
		ivalue, _ := StringToInt64(strvalue)
		ivalues = append(ivalues, ivalue)
	}
	return
}

func String(b []byte) (s string) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pstring.Data = pbytes.Data
	pstring.Len = pbytes.Len
	return
}

func Slice(s string) (b []byte) {
	pbytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pstring := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pbytes.Data = pstring.Data
	pbytes.Len = pstring.Len
	pbytes.Cap = pstring.Len
	return
}

func Version4ToInt(version string) int {
	arr := strings.Split(version, ".")
	fix_data := 1
	ver := 0
	for i := len(arr) - 1; i >= 0; i-- {
		d, _ := StringToInt(arr[i])
		ver = ver + d*fix_data
		fix_data *= 10000
	}
	return ver
}

// 获取保留n位小数的浮点型
func Round2(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}

// 判断分页是否末尾了
func IsListEnded(page, page_size, count, total int) (ended bool) {
	ended = true
	if page_size == count {
		if page*page_size < total {
			ended = false
		}
	}
	return
}

func IsEqual(f1, f2 float64) bool {
	return math.Abs(f1-f2) <= FLOAT_MIN
}

// 通过map主键唯一的特性过滤重复元素
func UniqueInt64Slice(slc []int64) []int64 {
	result := []int64{}
	tempMap := map[int64]bool{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = true
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 通过map主键唯一的特性过滤重复元素
func UniqueStringSlice(slc []string) []string {
	result := []string{}
	tempMap := map[string]bool{} // 存放不重复主键
	for _, e := range slc {
		if e == "" {
			continue
		}
		l := len(tempMap)
		tempMap[e] = true
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 通过map主键唯一的特性过滤重复元素
func UniqueIntSlice(slc []int) []int {
	result := []int{}
	tempMap := map[int]bool{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = true
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// struct2map
func StructToMap(v interface{}) (m map[string]interface{}) {

	m = make(map[string]interface{})
	bt, err := json.Marshal(v)
	if err != nil {
		return
	}

	json.Unmarshal(bt, &m)

	return
}

func GetCurDay() string {
	return time.Now().Format("2006-01-02")
}

// 手机号脱敏处理
func SensitiveTel(tel string) string {
	num := len(tel)
	if num >= 11 {
		tel = tel[:3] + "****" + tel[7:]
	} else if num > 5 {
		tel = tel[:2] + "***" + tel[5:]
	}
	return tel
}

func StringSliceToInt64Slice(vals []string) (vs []int64, err error) {
	vs = []int64{}
	if vals == nil {
		return
	}

	var n int64
	for _, v := range vals {
		if n, err = strconv.ParseInt(v, 10, 64); err != nil {
			return
		}
		vs = append(vs, n)
	}
	return
}
