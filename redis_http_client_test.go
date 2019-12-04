package gobaselib

import (
	// "fmt"
	"testing"
	"time"

	"gobaselib/cache"
)

func TestClient(t *testing.T) {
	redis_cfg := &cache.RedisConfig{}
	http, _ := NewRedisHttpClient(redis_cfg)

	timeout := time.Second * 2
	uri := "http://172.16.100.250:812/filminfo/list?getvideoinfo&ids=109852"
	headers := make(map[string]string)
	headers["X-Key"] = "fi:109852"
	res := http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	res = http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	headers["X-Key"] = "fihash:109852->video"
	res = http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	res = http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	headers["X-UseCacheOnFail"] = "1"
	res = http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

	uri = "http://127.0.0.1:812/filminfo/not/found"
	res = http.HttpGetJson(uri, headers, timeout, time.Second*10)
	t.Logf("res: %v, cached: %v, cost: %v", res.Ok, res.Cached, res.Stats.All)

}
