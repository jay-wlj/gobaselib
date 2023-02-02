package lock

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	redis "github.com/go-redis/redis/v8"
	"github.com/jay-wlj/gobaselib/log"
)

var (
	RedisLockKey = "redis_lock_key"
)

const (
	LOCK_WAIT_TIME_OUT = 999999 * time.Second
	unlockScript       = `
if redis.call("get",KEYS[1]) == ARGV[1] then
    return redis.call("del",KEYS[1])
else
    return 0
end`
)

type IRedisLock interface {
	Unlock(ctx context.Context, client redis.Cmdable) error
}

type redisLock struct {
	lockKey string
	lockId  string
}

func TryRedisLock(ctx context.Context, client redis.Cmdable, key string, timeout time.Duration) (bool, IRedisLock) {
	if client == nil {
		return false, nil
	}
	randNumber := gofakeit.UUID()
	curTime := time.Now()
	for {
		ok, err := client.SetNX(ctx, key, randNumber, timeout).Result()
		if err != nil {
			log.Errorf("TryLock SetNX %v fail! err=%v", key, err)
			return false, nil
		}
		if ok {
			// 获取到锁，返回
			return true, &redisLock{lockId: randNumber, lockKey: key}
		}
		// 锁超时判断
		if time.Now().Sub(curTime) > LOCK_WAIT_TIME_OUT {
			return false, nil
		}
		time.Sleep(10 * time.Millisecond)
	}

	return false, nil
}

func (t *redisLock) Unlock(ctx context.Context, client redis.Cmdable) error {
	script := redis.NewScript(unlockScript)
	_, err := script.Run(ctx, client, []string{t.lockKey}, t.lockId).Result()
	if err != nil {
		log.Errorf("Unlock %v fail! err=%v", t.lockKey, err)
		return err
	}
	return nil
}
