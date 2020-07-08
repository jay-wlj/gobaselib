package base

import (
	"reflect"
	"strings"
	"time"

	//"github.com/jay-wlj/pq"
	"github.com/jay-wlj/gobaselib/db"

	"github.com/shopspring/decimal"
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
		switch s.(type) {
		case decimal.Decimal, *decimal.Decimal: // 将不需要展开的结构体列出
			return s
		}
		out := make(map[string]interface{}, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)
			viewKey := field.Tag.Get("view")

			bOk := not_field && (tag_name != viewKey)
			bOk = bOk || (!not_field && tag_name == viewKey)
			bOk = bOk || (viewKey == "*")

			if !bOk {
				if field.Tag.Get("json") == "" && rv.Field(i).Kind() == reflect.Struct {
					bOk = true // 进入嵌套结构
				}
			}
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

					if !v.IsValid() {
						continue
					}
					// map,slice,function不能进行比较
					if v.Kind() != reflect.Map && v.Kind() != reflect.Slice && v.Kind() != reflect.Func {
						if v.Interface() == reflect.Zero(v.Type()).Interface() {
							continue
						}
					}
				}

				r := fliterObj(v, name)
				if r != nil {
					// 该成员为嵌套结构
					if jsonKey == "" {
						if v.Kind() == reflect.Struct {
							if m, ok := r.(map[string]interface{}); ok {
								for k, v := range m {
									out[k] = v
								}
							}
						}
					} else {
						out[jsonKey] = r
					}
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
		if v.CanInterface() {
			return SelectStructView(v.Interface(), name)
		}
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct && v.Elem().CanInterface() {
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

// 此方法为过滤掉或排除map中的key
func FilterStruct(s interface{}, include bool, fields ...string) interface{} {
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
	mfield := make(map[string]bool)
	for _, v := range fields {
		mfield[v] = true
	}
	// 拆解name 以"not "开头为排除拥有此标识的字段,否则只获取此标识的字段

	switch rv.Kind() {
	case reflect.Struct:
		switch s.(type) {
		case decimal.Decimal, *decimal.Decimal: // 将不需要展开的结构体列出
			return s
		case time.Time, *time.Time:
			return s
		case db.Jsonb:
			return s
		}
		out := make(map[string]interface{}, rt.NumField())
		for i := 0; i < rt.NumField(); i++ {
			field := rt.Field(i)

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

			if jsonKey != "" {
				if include && !mfield[jsonKey] { // 过滤掉没有在需要的字段里
					continue
				}
				if !include && mfield[jsonKey] { // 排除在里面的字段里
					continue
				}
			}

			v := rv.Field(i)
			// 判断字段是否为空 忽略
			if omitempty {
				if !v.IsValid() {
					continue
				}

				kt := v.Kind()
				if kt == reflect.Map || kt == reflect.Slice {
					if v.IsNil() || v.Len() == 0 {
						continue
					}
				} else if kt != reflect.Func { // map,slice,function不能进行比较
					if v.Interface() == reflect.Zero(v.Type()).Interface() {
						continue
					}
				}

				// TODO 这里获取key下面所有的结构体或数组
				if include && (kt == reflect.Slice || kt == reflect.Struct) {
					out[jsonKey] = v.Interface()
					continue
				}
			}

			r := fliterObjEx(v, include, fields...)
			if r != nil {
				// 该成员为嵌套结构
				if jsonKey == "" {
					if v.Kind() == reflect.Struct {
						if m, ok := r.(map[string]interface{}); ok {
							for k, v := range m {
								out[k] = v
							}
						}
					}
				} else {
					out[jsonKey] = r
				}
			}

		}
		return out
	case reflect.Slice:
		out := []interface{}{}
		for i := 0; i < rv.Len(); i++ {
			v := rv.Index(i)
			r := fliterObjEx(v, include, fields...)
			if r != nil {
				out = append(out, r)
			}
		}
		return out
	}

	return s
}

func fliterObjEx(v reflect.Value, include bool, fields ...string) interface{} {
	switch v.Kind() {
	case reflect.Struct:
		if v.CanInterface() {
			return FilterStruct(v.Interface(), include, fields...)
		}
	case reflect.Ptr:
		if v.Elem().Kind() == reflect.Struct && v.Elem().CanInterface() {
			return FilterStruct(v.Elem().Interface(), include, fields...)
		}
	case reflect.Slice:
		vs := []interface{}{}
		convert := true
		for j := 0; j < v.Len(); j++ {
			f := v.Index(j)
			kind := f.Kind()
			if kind == reflect.Ptr {
				f = f.Elem()
				kind = f.Kind()
			}
			// 对[]slice,struct才进行字段过滤
			if kind == reflect.Struct || kind == reflect.Slice {
				vs = append(vs, FilterStruct(f.Interface(), include, fields...))
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
	default:
		//fmt.Println(v.Kind())
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
