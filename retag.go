package gobaselib

import (
	"reflect"
	"strings"
)

// 此方法为过滤掉或排除view标签中的值,传入的s为struct或slice
func SelectStructView(s interface{}, name string) interface{} {
	if s == nil {
		return s
	}
	rt, rv := reflect.TypeOf(s), reflect.ValueOf(s)

	kind := rv.Kind()
	// 传进来的是结构体指针 则指向结构体
	if kind == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
		kind = rv.Kind()
	}
	if kind != reflect.Struct && kind != reflect.Slice {
		return s
	}

	// 拆解name 以"not "开头为排除拥有此标识的字段,否则只获取此标识的字段
	var not_field bool = false
	tag_name := name
	if strings.HasPrefix(name, "not ") {
		not_field = true
		tag_name = strings.Replace(name, "not ", "", -1)
	}

	switch rv.Kind() {
	case reflect.Struct:
		out := make(map[string]interface{}, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			viewKey := field.Tag.Get("view")

			bOk := not_field && (tag_name != viewKey)
			bOk = bOk || (!not_field && tag_name == viewKey)
			bOk = bOk || (viewKey == "*")

			// 获取view对应的name
			//if (viewKey == name) || viewKey == "*" {
			if bOk {
				jsonKey := field.Tag.Get("json")
				keys := strings.Split(jsonKey, ",") // 判断josn字段是否有其它信息 如 json:"key,omitempty"
				omitempty := false
				if len(keys) > 0 {
					omitempty = strings.Index(jsonKey, "omitempty") > 0
					jsonKey = keys[0] // 取字段别名
				}

				if jsonKey == "-" { // json标识没有导出 忽略此字段
					continue
				}

				v := rv.Field(i)
				// 判断字段是否为空 忽略
				if omitempty {
					if !v.IsValid() || v.Interface() == reflect.Zero(v.Type()).Interface() {
						continue
					}
				}

				r := fliterObj(v, name)
				if r != nil {
					out[jsonKey] = v.Interface()
				}
			}
		}
		return out
	case reflect.Slice:
		out := []interface{}{}
		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i)
			r := fliterObj(v, name)
			if r != nil {
				out = append(out, r)
			}
		}
		return out
	}

	return s
}

func fliterObj(v reflect.Value, name string) interface{} {
	switch v.Kind() {
	case reflect.Struct:
		return SelectStructView(v.Interface(), name)
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct {
			return SelectStructView(v.Elem().Interface(), name)
		}
	case reflect.Slice:
		vs := []interface{}{}
		convert := true
		for j := 0; j < v.Len(); j++ {
			kind := v.Index(j).Kind()

			// 对[]slice,struct才进行字段过滤
			if kind == reflect.Struct || kind == reflect.Slice {
				vs = append(vs, SelectStructView(v.Index(j).Interface(), name))
			} else {
				convert = false
				break
			}
		}
		if convert {
			return vs
		} else {
			return v.Interface()
		}
	}
	return v.Interface() // 返回原数据
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
