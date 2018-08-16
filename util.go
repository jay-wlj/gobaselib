package gobaselib

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"
	"math"
	"encoding/json"
)

const FLOAT_MIN = 0.0000001

func StringToInt(str string) (value int, err error) {
	value, err = strconv.Atoi(str)
	return value, err
}

func StringToInt64(str string) (value int64, err error) {
	value, err = strconv.ParseInt(str, 10, 64)
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

func Float64ToString(value float64) (strvalue string) {
	strvalue = strconv.FormatFloat(value, 'E', -1, 64)
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

func StringToIntSlice(str string, splite string) (ivalues []int) {
	strarr := strings.Split(str, splite)
	for _, strvalue := range strarr {
		ivalue, _ := StringToInt(strvalue)
		ivalues = append(ivalues, ivalue)
	}
	return
}

func StringToInt64Slice(str string, splite string) (ivalues []int64) {
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
	fix_data := 100000000
	ver := 0
	for _, num := range arr {
		d, _ := StringToInt(num)
		ver = ver + d*fix_data
		fix_data = fix_data / 100
	}
	return ver
}

// 判断分页是否末尾了
func IsListEnded(page, page_size, count, total int)(ended bool) {
	ended = true
	if page_size == count {
		if page*page_size < total {
			ended = false
		}
	}
	return
}

func IsEqual(f1, f2 float64) bool {
    return math.Abs(f1-f2) < FLOAT_MIN
}

// 通过map主键唯一的特性过滤重复元素
func UniqueInt64Slice(slc []int64) []int64 {
    result := []int64{}
    tempMap := map[int64]bool{}  // 存放不重复主键
    for _, e := range slc{
        l := len(tempMap)
        tempMap[e] = true
        if len(tempMap) != l{  // 加入map后，map长度变化，则元素不重复
            result = append(result, e)
        }
    }
    return result
}

// struct2map
func StructToMap(v interface{})map[string]interface{} {
	t := reflect.TypeOf(v)
	vf := reflect.ValueOf(v)
	m := make(map[string]interface{})
	if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
		str, err := json.Marshal(v)
		if err == nil {
			json.Unmarshal(str, &m)
		}
	} else {		
		for i:=0; i<t.NumField(); i++ {
			m[strings.ToLower(t.Field(i).Name)] = vf.Field(i).Interface()
		}
	}

	return m
}