package cache

import (
	"time"
	"github.com/jie123108/glog"
)

func NewRedisCacheFromCfg(rediserver, password, strTimeout string, dbindex int) (redis *RedisCache, err error) {
	cfg := RedisConfig{Addr:rediserver, Password:password, DBIndex:dbindex}
	cfg.Timeout, err = time.ParseDuration(strTimeout)
	if err != nil {
		glog.Error("invalid Config[timeout]: ", strTimeout)
		return
	}

	redis, err = NewRedisCache(&cfg)
	if err != nil {
		glog.Error("init redis failed! err:", err)
		return
	}
	glog.Info("init redis success!")
	return
}
