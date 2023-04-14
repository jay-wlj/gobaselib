package cache_redis

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/log"

	redis "github.com/go-redis/redis/v9"
)

var (
	ErrNotExist error
)

func init() {
	ErrNotExist = redis.Nil
}

type Config struct {
	Cli      redis.Cmdable `mapstructure:"-"` // 优先cli
	Network  string        `mapstructure:"network" toml:"network" json:"network,omitempty"`
	Nodes    []string      `mapstructure:"nodes" toml:"nodes" json:"nodes,omitempty"`
	Username string        `mapstructure:"username" toml:"username" json:"username,omitempty"`
	Password string        `mapstructure:"password" toml:"password" json:"password,omitempty"`
	DB       int           `mapstructure:"db" toml:"db" json:"db,omitempty"`
}

type redisCache struct {
	cacheName string
	redis.Cmdable
}

func NewClient(cfg *Config) (client *redisCache, err error) {
	client = &redisCache{}

	if cfg.Cli != nil {
		client.Cmdable = cfg.Cli
	} else {
		if len(cfg.Nodes) == 0 {
			return client, errors.New("no redis nodes")
		} else if len(cfg.Nodes) == 1 {
			client.Cmdable = redis.NewClient(&redis.Options{
				Network:  cfg.Network,
				Addr:     cfg.Nodes[0],
				Username: cfg.Username,
				Password: cfg.Password,
				DB:       cfg.DB,
			})
		} else {
			client.Cmdable = redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:    cfg.Nodes,
				Username: cfg.Username,
				Password: cfg.Password,
			})
		}
	}

	var pong string
	pong, err = client.Ping(context.TODO()).Result()
	if err != nil {
		log.Error("NewredisCache(", cfg.Nodes, ",", cfg.DB, ") failed! pong:", pong, " err:", err)
		client = nil
		return
	}

	defaultClient = client

	return client, nil
}

func (t *redisCache) Publish(ctx context.Context, channel, message string) (int64, error) {
	return t.Cmdable.Publish(ctx, channel, message).Result()
}

// 如果key为 xxx->field, 将从hash中获取数据.
func (t *redisCache) GetB(ctx context.Context, key string) (val []byte, err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		val, err = t.HGetB(ctx, keys[0], keys[1])
	} else {
		val, err = t.Cmdable.Get(ctx, key).Bytes()
	}
	return
}

// 如果key为 xxx->field, 将从hash中获取数据.
func (t *redisCache) Get(ctx context.Context, key string) (val string, err error) {
	b, err := t.GetB(ctx, key)
	if err == nil {
		val = string(b)
	}
	return
}

// 如果key为 xxx->field, 将数据存储到hash中.
func (t *redisCache) Set(ctx context.Context, key string, value interface{}, exptime time.Duration) (err error) {
	keys := strings.SplitN(key, "->", 2)
	if len(keys) == 2 {
		err = t.HSet(ctx, keys[0], keys[1], value, exptime)
	} else {
		err = t.Cmdable.Set(ctx, key, value, exptime).Err()
	}

	return
}

func (t *redisCache) HGet(ctx context.Context, key, field string) (val string, err error) {
	val, err = t.Cmdable.HGet(ctx, key, field).Result()
	return
}

func (t *redisCache) HGetI(ctx context.Context, key, field string) (val int64, err error) {
	val, err = t.Cmdable.HGet(ctx, key, field).Int64()
	return
}

func (t *redisCache) HGetB(ctx context.Context, key, field string) (val []byte, err error) {
	val, err = t.Cmdable.HGet(ctx, key, field).Bytes()
	return
}

func (t *redisCache) HGetF64(ctx context.Context, key, field string) (val float64, err error) {
	val, err = t.Cmdable.HGet(ctx, key, field).Float64()
	return
}

func (t *redisCache) HGetAll(ctx context.Context, key string) (val map[string]string, err error) {
	val, err = t.Cmdable.HGetAll(ctx, key).Result()
	return
}

func (t *redisCache) HSet(ctx context.Context, key, field string, value interface{}, exptime time.Duration) (err error) {
	err = t.Cmdable.HSet(ctx, key, field, value).Err()
	if err == nil && exptime > 0 {
		err = t.Cmdable.Expire(ctx, key, exptime).Err()
		if err != nil {
			log.Error("redis.Expire(", key, ", ", exptime, ") failed! err:", err)
			err = nil
		}
	}
	return
}

func (t *redisCache) HIncrBy(ctx context.Context, key, field string, incr int64) (n int64, err error) {
	n, err = t.Cmdable.HIncrBy(ctx, key, field, incr).Result()
	return
}
func (t *redisCache) HDel(ctx context.Context, key string, fields ...string) (n int64, err error) {
	n, err = t.Cmdable.HDel(ctx, key, fields...).Result()
	return
}

func (t *redisCache) SAddInt64(ctx context.Context, key string, ids []int64) (n int64, err error) {
	members := []interface{}{}
	for _, v := range ids {
		members = append(members, v)
	}
	n, err = t.Cmdable.SAdd(ctx, key, members...).Result()
	return
}

func (t *redisCache) SDiffInt64(ctx context.Context, key, key2 string) (vs []int64, err error) {
	var vals []string
	if vals, err = t.Cmdable.SDiff(ctx, key, key2).Result(); err == nil {
		vs, err = base.StringSliceToInt64Slice(vals)
	}
	return
}

func (t *redisCache) SMembersInt64(ctx context.Context, key string) (vs []int64, err error) {
	var vals []string
	if vals, err = t.Cmdable.SMembers(ctx, key).Result(); err == nil {
		vs, err = base.StringSliceToInt64Slice(vals)
	}
	return
}
func (t *redisCache) SRemInt64(ctx context.Context, key string, ids []int64) (n int64, err error) {
	members := []interface{}{}
	for _, v := range ids {
		members = append(members, v)
	}
	n, err = t.Cmdable.SRem(ctx, key, members...).Result()
	return
}

func (t *redisCache) Del(ctx context.Context, keys ...string) (n int64, err error) {
	n, err = t.Cmdable.Del(ctx, keys...).Result()
	return
}

func (t *redisCache) Eval(ctx context.Context, script string, keys []string, args ...interface{}) (n int64, err error) {
	v, err := t.Cmdable.Eval(ctx, script, keys, args...).Result()
	switch m := v.(type) {
	case int64:
		n = m
	case nil:
		// fmt.Printf("t is nil\n")
	default:
		fmt.Printf("unexpected type %T\n", t) // %T prints whatever type t has
	}

	return n, err
}

func (t *redisCache) LPush(ctx context.Context, key string, values ...interface{}) *redis.IntCmd {
	return t.Cmdable.LPush(ctx, key, values...)
}

func (t *redisCache) LRangeI64(ctx context.Context, key string, start, stop int64) (vs []int64, err error) {
	vs = []int64{}
	var rs []string
	rs, err = t.Cmdable.LRange(ctx, key, start, stop).Result()
	if err == nil && rs != nil {
		return base.StringSliceToInt64Slice(rs)
	}
	return
}

func (t *redisCache) LTrim(ctx context.Context, key string, start, end int64) *redis.StatusCmd {
	return t.Cmdable.LTrim(ctx, key, start, end)
}
func (t *redisCache) LLen(ctx context.Context, key string) (int64, error) {
	return t.Cmdable.LLen(ctx, key).Result()
}

func (t *redisCache) SetNx(ctx context.Context, key string, value interface{}, exptime time.Duration) (bool, error) {
	return t.Cmdable.SetNX(ctx, key, value, exptime).Result()
}

func (t *redisCache) HSetNx(ctx context.Context, key, field string, value interface{}) (bool, error) {
	return t.Cmdable.HSetNX(ctx, key, field, value).Result()
}

// 尝试获取锁
func (t *redisCache) TryLock(ctx context.Context, key string, exptime time.Duration) (ok bool, err error) {
	now := time.Now().UnixNano() // 当前时间ns
	acquire_time := now + int64(exptime)
	ok, err = t.Cmdable.SetNX(ctx, key, acquire_time, 0).Result()

	// 获取到锁或发生错误 直接返回
	if err != nil || ok {
		return
	}
	// 获取锁值是否过期
	var old int64
	if old, err = t.Cmdable.Get(ctx, key).Int64(); err != nil {
		return
	}
	if old < now {
		// 锁已过期 尝试获取锁 设置过期时间
		var new int64
		if new, err = t.Cmdable.GetSet(ctx, key, acquire_time).Int64(); err != nil {
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

func (t *redisCache) ZAddI64(ctx context.Context, key string, values []int64) (int64, error) {
	members := []*redis.Z{}
	for i, v := range values {
		members = append(members, &redis.Z{float64(i), v})
	}
	return t.Cmdable.ZAdd(ctx, key, members...).Result()
}

func (t *redisCache) ZRangeI64(ctx context.Context, key string, start, stop int64) (vs []int64, err error) {
	vs = []int64{}
	var rs []string
	rs, err = t.Cmdable.ZRange(ctx, key, start, stop).Result()
	if err == nil && rs != nil {
		return base.StringSliceToInt64Slice(rs)
	}
	return
}

func (t *redisCache) ZIsMember(ctx context.Context, key string, member string) (ok bool, err error) {
	ok = false
	_, err = t.Cmdable.ZRank(ctx, key, member).Result()
	if err == nil {
		ok = true
	}
	if err == redis.Nil {
		err = nil
	}
	return

}

// // type QueryFunc func(args... interface{})(val interface{}, err error)
// var errorType = reflect.TypeOf(make([]error, 1)).Elem()
// var stringType = reflect.TypeOf("")

// // in: t *redisCache, cachekey string, exptime time.Duration, query_func QueryFunc, args... interface{}
// //out: val interface{}, err error, cached string
// func CacheQuery(in []reflect.Value) []reflect.Value {
// 	t, _ := in[0].Interface().(*redisCache)

// 	cached := "miss"
// 	cachekey, _ := in[1].Interface().(string)

// 	exptime, _ := in[2].Interface().(time.Duration)
// 	query_func := in[3]
// 	args := in[4:]

// 	str, err := t.Get(context.TODO(), cachekey)
// 	if str != "" {
// 		//回调函数第一个返回值是个对象.
// 		ret_val_type := query_func.Type().Out(0)
// 		//动态创建一个对象, 用于接收json数据.
// 		val := reflect.New(ret_val_type)
// 		//Unmarshal需要interface类型.
// 		vali := val.Interface()
// 		err = json.Unmarshal([]byte(str), &vali)
// 		if err == nil {
// 			cached = "hit"
// 			// reflect.Indirect 将值得进行一次解引用.
// 			return []reflect.Value{reflect.Indirect(val), reflect.Zero(errorType), reflect.ValueOf(cached)}
// 		}
// 	}

// 	// 缓存中未查询到, 查询回调函数.
// 	values := query_func.Call(args)
// 	val := values[0].Interface()
// 	err, _ = values[1].Interface().(error)
// 	//查询成功, 缓存结果, 用于下次查询.
// 	if err == nil && val != nil {
// 		buf, err := json.Marshal(val)
// 		if err == nil {
// 			str = string(buf)
// 			t.Set(ctx, cachekey, str, exptime)
// 		}
// 	}

// 	values = append(values, reflect.ValueOf(cached))
// 	return values
// }

// func MakeCacheQuery(fptr interface{}) {
// 	fn := reflect.ValueOf(fptr).Elem()
// 	v := reflect.MakeFunc(fn.Type(), CacheQuery)
// 	fn.Set(v)
// }

// // in: t *redisCache, hashcachekey, field string, exptime time.Duration, query_func QueryFunc, args... interface{}
// //out: val interface{}, err error, cached string
// func HCacheQuery(in []reflect.Value) []reflect.Value {
// 	t, _ := in[0].Interface().(*redisCache)

// 	cached := "miss"
// 	cachekey, _ := in[1].Interface().(string)
// 	field, _ := in[2].Interface().(string)

// 	exptime, _ := in[3].Interface().(time.Duration)
// 	query_func := in[4]
// 	args := in[5:]

// 	str, err := t.HGet(cachekey, field)
// 	if str != "" {
// 		//回调函数第一个返回值是个对象.
// 		ret_val_type := query_func.Type().Out(0)
// 		//动态创建一个对象, 用于接收json数据.
// 		val := reflect.New(ret_val_type)
// 		//Unmarshal需要interface类型.
// 		vali := val.Interface()
// 		err = json.Unmarshal([]byte(str), &vali)
// 		if err == nil {
// 			cached = "hit"
// 			// reflect.Indirect 将值得进行一次解引用.
// 			return []reflect.Value{reflect.Indirect(val), reflect.Zero(errorType), reflect.ValueOf(cached)}
// 		}
// 	}

// 	// 缓存中未查询到, 查询回调函数.
// 	values := query_func.Call(args)
// 	val := values[0].Interface()
// 	err, _ = values[1].Interface().(error)
// 	//查询成功, 缓存结果, 用于下次查询.
// 	if err == nil && val != nil {
// 		buf, err := json.Marshal(val)
// 		if err == nil {
// 			str = string(buf)
// 			t.HSet(cachekey, field, str, exptime)
// 		}
// 	}

// 	values = append(values, reflect.ValueOf(cached))
// 	return values
// }
// func MakeHCacheQuery(fptr interface{}) {
// 	fn := reflect.ValueOf(fptr).Elem()
// 	v := reflect.MakeFunc(fn.Type(), HCacheQuery)
// 	fn.Set(v)
// }
