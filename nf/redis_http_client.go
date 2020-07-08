package base

import (
	// "fmt"
	"encoding/json"
	"time"

	"github.com/jay-wlj/gobaselib/cache"

	"github.com/jie123108/glog"
)

type RedisHttpClient struct {
	redis_cfg *cache.RedisConfig // redis配置.
	cache     *cache.RedisCache
}

type CachedResp struct {
	Resp
	Uri        string
	UpdateTime int32 //unixtime
	Exptime    int32 //unixtime
}

func NewRedisHttpClient(redis_cfg *cache.RedisConfig) (client *RedisHttpClient, err error) {
	if redis_cfg.Timeout == 0 {
		redis_cfg.Timeout = time.Hour * 24 * 2
	}

	cacheutil, err1 := NewCacheUtil(redis_cfg)
	if err1 != nil {
		err = err1
		return
	}

	client = &RedisHttpClient{redis_cfg, cacheutil.RedisCache}

	return
}

func (this *RedisHttpClient) HttpGetJson(uri string, headers map[string]string,
	timeout time.Duration, exptime time.Duration) *OkJson {
	key, ok := headers["X-Key"]
	if !ok {
		res := HttpGetJson(uri, headers, timeout)
		return res
	}

	res_cached := &OkJson{}
	now := time.Now()
	body_len := 0
	str, err := this.cache.Get(key)
	defer write_debug_ok_json(now, res_cached, &body_len)
	if err == nil && str != "" {
		var data CachedResp
		err = json.Unmarshal([]byte(str), &data)
		// exptime > now 没过期.
		if err == nil {
			res_cached.StatusCode = data.StatusCode
			res_cached.RawBody = data.RawBody
			res_cached.Headers = data.Headers
			res_cached.ReqDebug = data.ReqDebug
			_, ok = headers["X-UseCacheOnFail"]
			if !ok && data.Exptime > int32(now.Unix()) {
				return res_cached.okJsonParse()
			}
		}
	}

	res := HttpGetJson(uri, headers, timeout)

	if res.StatusCode == 200 {
		data := &CachedResp{}
		data.StatusCode = res.StatusCode
		data.RawBody = res.RawBody
		data.Headers = res.Headers
		data.ReqDebug = res.ReqDebug
		data.Uri = uri
		data.UpdateTime = int32(now.Unix())
		data.Exptime = int32(now.Add(exptime).Unix())
		bt, _ := json.Marshal(data)
		str = string(bt)
		err = this.cache.Set(key, str, this.redis_cfg.Timeout)
		if err != nil {
			glog.Errorf("cache.Set('%s', '%v', '%v') failed! err:%v", key, data, this.redis_cfg.Timeout, err)
		}
	} else {
		if res_cached.StatusCode > 0 {
			return res_cached.okJsonParse()
		}
	}

	return res
}
