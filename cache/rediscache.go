package cache

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jie123108/glog"
	redis "gopkg.in/redis.v5"
)

var (
	ErrNotExist error
)

func init() {
	ErrNotExist = redis.Nil
}

type RedisConfig struct {
	CacheName string
	Addr      string // host:port
	Password  string
	DBIndex   int
	Timeout   time.Duration
}

type RedisCache struct {
	cacheName string
	Cfg       *RedisConfig
	client    *redis.Client
}

func NewRedisCache(cfg *RedisConfig) (cache *RedisCache, err error) {
	cache = &RedisCache{}
	cache.cacheName = cfg.CacheName
	if cache.cacheName == "" {
		cache.cacheName = fmt.Sprintf("redis_%s", cfg.Addr)
	}
	if cfg.Addr == "" {
		cfg.Addr = "127.0.0.1:6379"
	}
	cache.Cfg = cfg

	cache.client = redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DBIndex})
	pong, err := cache.client.Ping().Result()
	if err != nil {
		glog.Error("NewRedisCache(", cfg.Addr, ",", cfg.DBIndex, ") failed! pong:", pong, " err:", err)
		cache = nil
		return
	}

	return
}

func (this *RedisCache) Subscribe(channel ...string) (pub *redis.PubSub, err error) {
	pub, err = this.client.Subscribe(channel...)
	return
}

func (this *RedisCache) Publish(channel, message string) (int64, error) {
	return this.client.Publish(channel, message).Result()
}

// 如果key为 xxx->field, 将从hash中获取数据.
func (this *RedisCache) GetB(key string) (val []byte, err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		val, err = this.HGetB(keys[0], keys[1])
	} else {
		val, err = this.client.Get(key).Bytes()
	}
	return
}

// 如果key为 xxx->field, 将从hash中获取数据.
func (this *RedisCache) Get(key string) (val string, err error) {
	b, err := this.GetB(key)
	if err == nil {
		val = string(b)
	}
	return
}

// 如果key为 xxx->field, 将数据存储到hash中.
func (this *RedisCache) Set(key string, value interface{}, exptime time.Duration) (err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		err = this.HSet(keys[0], keys[1], value, exptime)
	} else {
		err = this.client.Set(key, value, exptime).Err()
	}
	return
}

func (this *RedisCache) HGet(key, field string) (val string, err error) {
	val, err = this.client.HGet(key, field).Result()
	return
}

func (this *RedisCache) HGetI(key, field string) (val int64, err error) {
	val, err = this.client.HGet(key, field).Int64()
	return
}

func (this *RedisCache) HGetB(key, field string) (val []byte, err error) {
	val, err = this.client.HGet(key, field).Bytes()
	return
}

func (this *RedisCache) HGetF64(key, field string) (val float64, err error) {
	val, err = this.client.HGet(key, field).Float64()
	return
}

func (this *RedisCache) HGetAll(key string) (val map[string]string, err error) {
	val, err = this.client.HGetAll(key).Result()
	return
}

func (this *RedisCache) HSet(key, field string, value interface{}, exptime time.Duration) (err error) {
	err = this.client.HSet(key, field, value).Err()
	if err == nil && exptime > 0 {
		err = this.client.Expire(key, exptime).Err()
		if err != nil {
			glog.Error("redis.Expire(", key, ", ", exptime, ") failed! err:", err)
			err = nil
		}
	}
	return
}

func (this *RedisCache) HIncrBy(key, field string, incr int64) (n int64, err error) {
	n, err = this.client.HIncrBy(key, field, incr).Result()
	return
}
func (this *RedisCache) HDel(key string, fields ...string) (n int64, err error) {
	n, err = this.client.HDel(key, fields...).Result()
	return
}

func (this *RedisCache) SAdd(key string, members ...interface{}) (n int64, err error) {
	n, err = this.client.SAdd(key, members...).Result()
	return
}
func (this *RedisCache) Del(keys ...string) (n int64, err error) {
	n, err = this.client.Del(keys...).Result()
	return
}

func (this *RedisCache) Eval(script string, keys []string, args ...interface{}) (n int64, err error) {
	t, err := this.client.Eval(script, keys, args...).Result()
	switch t := t.(type) {
	case int64:
		n = t
	case nil:
		// fmt.Printf("t is nil\n")
	default:
		fmt.Printf("unexpected type %T\n", t) // %T prints whatever type t has
	}

	return n, err
}

// type QueryFunc func(args... interface{})(val interface{}, err error)
var errorType = reflect.TypeOf(make([]error, 1)).Elem()
var stringType = reflect.TypeOf("")

// in: this *RedisCache, cachekey string, exptime time.Duration, query_func QueryFunc, args... interface{}
//out: val interface{}, err error, cached string
func CacheQuery(in []reflect.Value) []reflect.Value {
	this, _ := in[0].Interface().(*RedisCache)

	cached := "miss"
	cachekey, _ := in[1].Interface().(string)

	exptime, _ := in[2].Interface().(time.Duration)
	query_func := in[3]
	args := in[4:]

	str, err := this.Get(cachekey)
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
			this.Set(cachekey, str, exptime)
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

// in: this *RedisCache, hashcachekey, field string, exptime time.Duration, query_func QueryFunc, args... interface{}
//out: val interface{}, err error, cached string
func HCacheQuery(in []reflect.Value) []reflect.Value {
	this, _ := in[0].Interface().(*RedisCache)

	cached := "miss"
	cachekey, _ := in[1].Interface().(string)
	field, _ := in[2].Interface().(string)

	exptime, _ := in[3].Interface().(time.Duration)
	query_func := in[4]
	args := in[5:]

	str, err := this.HGet(cachekey, field)
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
			this.HSet(cachekey, field, str, exptime)
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

func (this *RedisCache) CacheStats() *redis.PoolStats {
	return this.client.PoolStats()
}

func (this *RedisCache) LPush(key string, values ...interface{}) *redis.IntCmd {
	return this.client.LPush(key, values...)
}

func (this *RedisCache) LRange(key string, start, stop int64) ([]string, error) {
	return this.client.LRange(key, start, stop).Result()
}

func (this *RedisCache) LTrim(key string, start, end int64) *redis.StatusCmd {
	return this.client.LTrim(key, start, end)
}

func (this *RedisCache) SetNx(key string, value interface{}, exptime time.Duration) (bool, error) {
	return this.client.SetNX(key, value, exptime).Result()
}

func (this *RedisCache) HSetNx(key, field string, value interface{}) (bool, error) {
	return this.client.HSetNX(key, field, value).Result()
}
