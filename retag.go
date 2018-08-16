package gobaselib

import (
	"reflect"
)

func SelectStructView(s interface{}, name string) map[string]interface{} {
	rt, rv := reflect.TypeOf(s), reflect.ValueOf(s)
	out := make(map[string]interface{}, rt.NumField())
	for i := 0; i <rt.NumField(); i++ {
		field := rt.Field(i)
		viewKey := field.Tag.Get("view")
		// 获取view对应的name
		if viewKey ==  name || viewKey == "*" {
			jsonKey := field.Tag.Get("json")
			v := rv.Field(i)
			switch v.Kind() {
			case reflect.Struct:
				out[jsonKey] = SelectStructView(v.Interface(), name)
				continue
			case reflect.Slice:
				vs := []interface{}{}
				convert := true
				for j:=0; j<v.Len(); j++ {
					// 对[]struct才进行字段过滤
					if v.Index(j).Kind() == reflect.Struct {
						vs = append(vs, SelectStructView(v.Index(j).Interface(), name))
					} else {
						convert = false
						break
					}
				}
				if convert {
					out[jsonKey] = vs
				} else {
					out[jsonKey] = v.Interface()
				}		
				continue
			default:
				break
			}				
			out[jsonKey] = v.Interface()
		}		
	}
	return out
}

func SelectStructFileds(s interface{}, tag, name string) map[string]interface{} {
	rt, rv := reflect.TypeOf(s), reflect.ValueOf(s)
	out := make(map[string]interface{}, rt.NumField())
	for i := 0; i <rt.NumField(); i++ {
		field := rt.Field(i)
		TagKey := field.Tag.Get("tag")
		if TagKey ==  name {
			jsonKey := field.Tag.Get("json")
			out[jsonKey] = rv.Field(i).Interface()
		}
	}
	return out
}