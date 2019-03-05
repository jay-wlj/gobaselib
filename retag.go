package gobaselib

import (
	"reflect"
	"strings"
)

func SelectStructView(s interface{}, name string) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	rt, rv := reflect.TypeOf(s), reflect.ValueOf(s)
	// 传进来的是结构体指针 则指向结构体
	if rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Struct {
		rt = rt.Elem()
		rv = rv.Elem()
	}
	out := make(map[string]interface{}, rt.NumField())

	// 拆解name 以"not "开头为排除拥有此标识的字段,否则只获取此标识的字段
	var not_field bool = false
	tag_name := name
	if strings.HasPrefix(name, "not ") {
		not_field = true
		tag_name = strings.Replace(name, "not ", "", -1)
	}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		viewKey := field.Tag.Get("view")

		bOk := not_field && (tag_name != viewKey)
		bOk = bOk || (!not_field && tag_name == viewKey)
		// 获取view对应的name
		//if (viewKey == name) || viewKey == "*" {
		if bOk || viewKey == "*" {
			jsonKey := field.Tag.Get("json")
			keys := strings.Split(jsonKey, ",") // 判断josn字段是否有其它信息 如 json:"key,omitempty"
			if len(keys) > 0 {
				jsonKey = keys[0]
			}
			if jsonKey == "-" { // json标识没有导出 忽略此字段
				continue
			}

			v := rv.Field(i)
			switch v.Kind() {
			case reflect.Struct:
				out[jsonKey] = SelectStructView(v.Interface(), name)
				continue
			case reflect.Ptr:
				if v.Elem().Kind() == reflect.Struct {
					out[jsonKey] = SelectStructView(v.Elem().Interface(), name)
					continue
				}
			case reflect.Slice:
				vs := []interface{}{}
				convert := true
				for j := 0; j < v.Len(); j++ {
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
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		TagKey := field.Tag.Get(tag)
		if TagKey == name || TagKey == "*" {
			jsonKey := field.Tag.Get("json")
			out[jsonKey] = rv.Field(i).Interface()
		}
	}
	return out
}
