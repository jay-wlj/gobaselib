package cache

import (
	"fmt"
	"gobaselib/log"
	//_ "github.com/jinzhu/gorm/dialects/postgres"
)

var g_rediscache *RedisCache
var g_apicache *RedisCache

type RedisCfg struct {
	Master RedisConfig  `json:"master"`
	Slave  *RedisConfig `json:"slave"`
}

var m_masterRedis map[string]*RedisCache
var m_slaveRedis map[string]*RedisCache

func init() {
	m_masterRedis = make(map[string]*RedisCache)
	m_slaveRedis = make(map[string]*RedisCache)
}
func InitRedis(vs map[string]RedisCfg) (err error) {
	for k, v := range vs {
		var ch *RedisCache
		if ch, err = NewRedisCache(&v.Master); err != nil {
			panic(fmt.Sprintf("InitRedis fail! cfg=", v.Master, " err=", err))
		}
		m_masterRedis[k] = ch

		if v.Slave != nil {
			if ch, err = NewRedisCache(v.Slave); err != nil {
				panic(fmt.Sprintf("InitRedis fail! cfg=", v.Slave, " err=", err))
			}
			m_slaveRedis[k] = ch
		}
		log.Info("InitRedis success! redis:", v.Master)
	}
	return
}

// type Reader interface {
// 	HGet(key string, field string) *redis.StringCmd
// 	HGetAll(key string) *redis.StringStringMapCmd
// 	Get(key string, field string) *redis.StringCmd
// 	Exists(key string) bool
// }

// type Writer interface {
// 	HSet(key string, field string, val int64) *redis.BoolCmd
// 	HIncrBy(key string, field string, incr int64) *redis.IntCmd
// 	SAdd(key string, members ...interface{}) *redis.IntCmd
// 	Set(key string, val string) *redis.StatusCmd
// 	Get(key string, field string) *redis.StringCmd
// 	Exists(key string) bool
// }

func GetWriter(key string) (*RedisCache, error) {
	if v, ok := m_masterRedis[key]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("GetWriter(): redis writer is unvaild! key=", key)
}
func GetReader(key string) (*RedisCache, error) {
	if v, ok := m_slaveRedis[key]; ok {
		return v, nil
	}

	// 没有在只读的redis 则读取master的redis
	return GetWriter(key)

	return nil, fmt.Errorf("GetReader(): redis reader is unvaild! key=", key)
}

func NewRedisCacheFromCfg(rediserver, password, strTimeout string, dbindex int) (redis *RedisCache, err error) {
	cfg := RedisConfig{Addr: rediserver, Password: password, DBIndex: dbindex, TimeoutStr: strTimeout}
	redis, err = NewRedisCache(&cfg)
	if err != nil {
		log.Error("init redis failed! err:", err)
		return
	}
	log.Info("init redis success!")
	return
}
