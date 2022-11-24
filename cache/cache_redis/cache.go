package cache_redis

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"time"

	"github.com/jay-wlj/gobaselib/log"
)

var defaultClient *redisCache

func Init(cfg *Config) error {
	if cfg == nil {
		log.Warnf("empty config")
		return errors.New("empty config")
	}

	var err error
	defaultClient, err = NewClient(cfg)
	if err != nil {
		log.Warn("redis configuration error")
		return err
	}

	return nil
}

func DefaultClient() *redisCache {
	return defaultClient
}

// type QueryFunc func(args... interface{})(val interface{}, err error)
var errorType = reflect.TypeOf(make([]error, 1)).Elem()
var stringType = reflect.TypeOf("")

// in: this *RedisCache, cachekey string, exptime time.Duration, query_func QueryFunc, ctx context.Context, args... interface{}
//out: val interface{}, err error, cached string
func CacheQuery(in []reflect.Value) []reflect.Value {
	this, _ := in[0].Interface().(*redisCache)

	cached := "miss"
	cachekey, _ := in[1].Interface().(string)

	exptime, _ := in[2].Interface().(time.Duration)
	query_func := in[3]
	args := in[4:]
	ctx, _ := args[0].Interface().(context.Context)

	str, err := this.Get(ctx, cachekey)
	if str != "" {
		//回调函数第一个返回值是个对象.
		ret_val_type := query_func.Type().Out(0)
		//动态创建一个对象, 用于接收json数据.
		val := reflect.New(ret_val_type)
		//Unmarshal需要interface类型.
		vali := val.Interface()
		err = json.Unmarshal([]byte(str), &vali)
		if err == nil {
			cached = "hit"
			// reflect.Indirect 将值得进行一次解引用.
			return []reflect.Value{reflect.Indirect(val), reflect.Zero(errorType), reflect.ValueOf(cached)}
		}
	}

	// 缓存中未查询到, 查询回调函数.
	values := query_func.Call(args)
	val := values[0].Interface()
	err, _ = values[1].Interface().(error)
	//查询成功, 缓存结果, 用于下次查询.
	if err == nil && val != nil {
		buf, err := json.Marshal(val)
		if err == nil {
			str = string(buf)
			this.Set(ctx, cachekey, str, exptime)
		}
	}

	values = append(values, reflect.ValueOf(cached))
	return values
}

func MakeCacheQuery(fptr interface{}) {
	fn := reflect.ValueOf(fptr).Elem()
	v := reflect.MakeFunc(fn.Type(), CacheQuery)
	fn.Set(v)
}

// in: this *RedisCache, hashcachekey, field string, exptime time.Duration, query_func QueryFunc, ctx context.Context, args... interface{}
//out: val interface{}, err error, cached string
func HCacheQuery(in []reflect.Value) []reflect.Value {
	this, _ := in[0].Interface().(*redisCache)

	cached := "miss"
	cachekey, _ := in[1].Interface().(string)
	field, _ := in[2].Interface().(string)

	exptime, _ := in[3].Interface().(time.Duration)
	query_func := in[4]
	args := in[5:]
	ctx, _ := args[0].Interface().(context.Context)

	str, err := this.HGet(ctx, cachekey, field)
	if str != "" {
		//回调函数第一个返回值是个对象.
		ret_val_type := query_func.Type().Out(0)
		//动态创建一个对象, 用于接收json数据.
		val := reflect.New(ret_val_type)
		//Unmarshal需要interface类型.
		vali := val.Interface()
		err = json.Unmarshal([]byte(str), &vali)
		if err == nil {
			cached = "hit"
			// reflect.Indirect 将值得进行一次解引用.
			return []reflect.Value{reflect.Indirect(val), reflect.Zero(errorType), reflect.ValueOf(cached)}
		}
	}

	// 缓存中未查询到, 查询回调函数.
	values := query_func.Call(args)
	val := values[0].Interface()
	err, _ = values[1].Interface().(error)
	//查询成功, 缓存结果, 用于下次查询.
	if err == nil && val != nil {
		buf, err := json.Marshal(val)
		if err == nil {
			str = string(buf)
			this.HSet(ctx, cachekey, field, str, exptime)
		}
	}

	values = append(values, reflect.ValueOf(cached))
	return values
}
func MakeHCacheQuery(fptr interface{}) {
	fn := reflect.ValueOf(fptr).Elem()
	v := reflect.MakeFunc(fn.Type(), HCacheQuery)
	fn.Set(v)
}
