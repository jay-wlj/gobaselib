package cache

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
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
	CacheName  string
	Addr       string // host:port
	Password   string
	DBIndex    int           `yaml:"dbindex"`
	Timeout    time.Duration `yaml:"-"`
	TimeoutStr string        `yaml:"timeout"`
}

type RedisCache struct {
	cacheName string
	Cfg       *RedisConfig
	*redis.Client
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

	if cfg.Timeout == 0 && cfg.TimeoutStr != "" {
		if cfg.Timeout, err = time.ParseDuration(cfg.TimeoutStr); err != nil {
			glog.Error("NewRedisCache(", cfg.Addr, ",", cfg.DBIndex, ") failed! timeout:", cfg.Timeout, " err:", err)
			cache = nil
			return
		}
	}

	cache.Client = redis.NewClient(&redis.Options{Addr: cfg.Addr, Password: cfg.Password, DB: cfg.DBIndex, DialTimeout: cfg.Timeout})
	var pong string
	pong, err = cache.Client.Ping().Result()
	if err != nil {
		glog.Error("NewRedisCache(", cfg.Addr, ",", cfg.DBIndex, ") failed! pong:", pong, " err:", err)
		cache = nil
		return
	}
	return
}

func (this *RedisCache) Subscribe(channel ...string) (pub *redis.PubSub, err error) {

	pub, err = this.Client.Subscribe(channel...)
	return
}

func (this *RedisCache) Publish(channel, message string) (int64, error) {
	return this.Client.Publish(channel, message).Result()
}

// 如果key为 xxx->field, 将从hash中获取数据.
func (this *RedisCache) GetB(key string) (val []byte, err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		val, err = this.HGetB(keys[0], keys[1])
	} else {
		val, err = this.Client.Get(key).Bytes()
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
		err = this.Client.Set(key, value, exptime).Err()
	}

	return
}

func (this *RedisCache) HGet(key, field string) (val string, err error) {
	val, err = this.Client.HGet(key, field).Result()
	return
}

func (this *RedisCache) HGetI(key, field string) (val int64, err error) {
	val, err = this.Client.HGet(key, field).Int64()
	return
}

func (this *RedisCache) HGetB(key, field string) (val []byte, err error) {
	val, err = this.Client.HGet(key, field).Bytes()
	return
}

func (this *RedisCache) HGetF64(key, field string) (val float64, err error) {
	val, err = this.Client.HGet(key, field).Float64()
	return
}

func (this *RedisCache) HGetAll(key string) (val map[string]string, err error) {
	val, err = this.Client.HGetAll(key).Result()
	return
}

func (this *RedisCache) HSet(key, field string, value interface{}, exptime time.Duration) (err error) {
	err = this.Client.HSet(key, field, value).Err()
	if err == nil && exptime > 0 {
		err = this.Client.Expire(key, exptime).Err()
		if err != nil {
			glog.Error("redis.Expire(", key, ", ", exptime, ") failed! err:", err)
			err = nil
		}
	}
	return
}

func (this *RedisCache) HIncrBy(key, field string, incr int64) (n int64, err error) {
	n, err = this.Client.HIncrBy(key, field, incr).Result()
	return
}
func (this *RedisCache) HDel(key string, fields ...string) (n int64, err error) {
	n, err = this.Client.HDel(key, fields...).Result()
	return
}

func (this *RedisCache) SAddInt64(key string, ids []int64) (n int64, err error) {
	members := []interface{}{}
	for _, v := range ids {
		members = append(members, v)
	}
	n, err = this.Client.SAdd(key, members...).Result()
	return
}

func (this *RedisCache) SMembersInt64(key string) (vs []int64, err error) {
	var vals []string
	if vals, err = this.Client.SMembers(key).Result(); err == nil {
		vs, err = StringSliceToInt64Slice(vals)
	}
	return
}
func (this *RedisCache) SRemInt64(key string, ids []int64) (n int64, err error) {
	members := []interface{}{}
	for _, v := range ids {
		members = append(members, v)
	}
	n, err = this.Client.SRem(key, members...).Result()
	return
}

func (this *RedisCache) Del(keys ...string) (n int64, err error) {
	n, err = this.Client.Del(keys...).Result()
	return
}

func (this *RedisCache) Eval(script string, keys []string, args ...interface{}) (n int64, err error) {
	t, err := this.Client.Eval(script, keys, args...).Result()
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
	return this.Client.PoolStats()
}

func (this *RedisCache) LPush(key string, values ...interface{}) *redis.IntCmd {
	return this.Client.LPush(key, values...)
}

func (this *RedisCache) LRangeI64(key string, start, stop int64) (vs []int64, err error) {
	vs = []int64{}
	var rs []string
	rs, err = this.Client.LRange(key, start, stop).Result()
	if err == nil && rs != nil {
		return StringSliceToInt64Slice(rs)
	}
	return
}

func (this *RedisCache) LTrim(key string, start, end int64) *redis.StatusCmd {
	return this.Client.LTrim(key, start, end)
}
func (this *RedisCache) LLen(key string) (int64, error) {
	return this.Client.LLen(key).Result()
}

func (this *RedisCache) SetNx(key string, value interface{}, exptime time.Duration) (bool, error) {
	return this.Client.SetNX(key, value, exptime).Result()
}

func (this *RedisCache) HSetNx(key, field string, value interface{}) (bool, error) {
	return this.Client.HSetNX(key, field, value).Result()
}

// 尝试获取锁
func (this *RedisCache) TryLock(key string, exptime time.Duration) (ok bool, err error) {
	now := time.Now().UnixNano() // 当前时间ns
	acquire_time := now + int64(exptime)
	ok, err = this.Client.SetNX(key, acquire_time, 0).Result()

	// 获取到锁或发生错误 直接返回
	if err != nil || ok {
		return
	}
	// 获取锁值是否过期
	var old int64
	if old, err = this.Client.Get(key).Int64(); err != nil {
		return
	}
	if old < now {
		// 锁已过期 尝试获取锁 设置过期时间
		var new int64
		if new, err = this.Client.GetSet(key, acquire_time).Int64(); err != nil {
			return
		}
		if old == new {
			// 成功获取到锁
			ok = true
			return
		} else {
			// 锁已被其它线程获取
		}
	}
	return
}

func (this *RedisCache) ZAddI64(key string, values []int64) (int64, error) {
	members := []redis.Z{}
	for i, v := range values {
		members = append(members, redis.Z{float64(i), v})
	}
	return this.Client.ZAdd(key, members...).Result()
}

func (this *RedisCache) ZRangeI64(key string, start, stop int64) (vs []int64, err error) {
	vs = []int64{}
	var rs []string
	rs, err = this.Client.ZRange(key, start, stop).Result()
	if err == nil && rs != nil {
		return StringSliceToInt64Slice(rs)
	}
	return
}

func (this *RedisCache) ZIsMember(key string, member string) (ok bool, err error) {
	ok = false
	_, err = this.Client.ZRank(key, member).Result()
	if err == nil {
		ok = true
	}
	if err == redis.Nil {
		err = nil
	}
	return

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
