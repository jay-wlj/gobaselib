package mt

import (
	// "fmt"
	base "gobaselib"
	"gobaselib/cache"
	// "github.com/jie123108/glog"
	"testing"
	"time"
)

func init() {
	// method, uri, args, headers, body_encrypt

}

func TestCachedHttpGet(t *testing.T) {
	redis_cfg := &cache.RedisConfig{}
	client, _ := base.NewRedisHttpClient(redis_cfg)
	uri := "http://172.16.100.251:83/v1/api/filminfo/detail?film_id=109852"
	app_key := "786f0897555b057037aa44714890260b"
	headers := make(map[string]string)
	headers["Host"] = "f-api.nicefilm.com"
	headers["X-Mt-AppId"] = "api_v2"
	headers["X-Mt-rid"] = "1"
	headers["X-Mt-Platform"] = "test"
	headers["X-Mt-Version"] = "1.5.0"
	headers["X-Key"] = "fid:109852"
	res := CachedNfHttpGet(client, time.Second*50, uri, headers, time.Second*1, app_key)
	t.Logf("111 ok: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	res = CachedNfHttpGet(client, time.Second*50, uri, headers, time.Second*1, app_key)
	t.Logf("222 ok: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)
}
