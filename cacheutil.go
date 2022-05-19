package base

import (
	"encoding/json"
	"github.com/jay-wlj/gobaselib/log"
	"sync"
	"time"

	"github.com/jay-wlj/gobaselib/cache"
)

type QueryFunc func(value interface{}, args ...interface{}) error

type CacheUtil struct {
	RedisConfig *cache.RedisConfig
	RedisCache  *cache.RedisCache
}

var Map_rediscache map[string]*CacheUtil
var G_redsclient sync.RWMutex

func NewCacheUtil(redisconfig *cache.RedisConfig) (*CacheUtil, error) {
	G_redsclient.Lock()
	defer G_redsclient.Unlock()

	if Map_rediscache == nil {
		log.Errorf("-----------------NewCacheUtil host:%v", redisconfig)
		Map_rediscache = make(map[string]*CacheUtil)
	}

	value := Map_rediscache[redisconfig.Addr]
	if nil == value {
		client, err := cache.NewRedisCache(redisconfig)
		if err != nil {
			log.Errorf("NewRedisCache(%v) failed! err:%v", redisconfig, err)
			return nil, err
		}
		Map_rediscache[redisconfig.Addr] = new(CacheUtil)
		Map_rediscache[redisconfig.Addr].RedisConfig = redisconfig
		Map_rediscache[redisconfig.Addr].RedisCache = client
	}
	return Map_rediscache[redisconfig.Addr], nil
}

func (this *CacheUtil) GetCache(cachename string, key string, value interface{}) error {
	buf, err := this.RedisCache.Get(key)
	if err == nil {
		err = json.Unmarshal(Slice(buf), value)
	}
	return err
}

func (this *CacheUtil) SetCache(cachename string, key string, value interface{}, exptime time.Duration) error {
	buf, err := json.Marshal(value)
	err = this.RedisCache.Set(key, String(buf), exptime)
	if err != nil {
		log.Errorf("client.Set(%v, %v) failed! err:%v", key, value, err)
		return err
	}

	return err
}

func (this *CacheUtil) DeleteCache(cachename string, key string) error {
	num, err := this.RedisCache.Del(key)
	if err == nil {
		log.Infof("deletecache:%v %v %v succ", cachename, key, num)
	} else {
		log.Errorf("deletecache:%v %v failed! err:%v", cachename, key, err)
	}
	return err
}

func (this *CacheUtil) CacheQuery(queryfunc QueryFunc, cachename string, key string, value interface{}, exptime time.Duration, args ...interface{}) error {
	err := this.GetCache(cachename, key, value)
	if err == nil {
		return err
	}
	err = queryfunc(value, args...)
	if err == nil {
		err = this.SetCache(cachename, key, value, exptime)
		if err != nil {
			log.Errorf("client.SetCache(%v,%v) failed! err:%v", key, value, err)
		}
	}
	return nil
}
