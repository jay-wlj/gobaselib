package gobaselib

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

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
