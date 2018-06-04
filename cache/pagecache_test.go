package cache

import (
    "testing"
    "math"
)

var pct_cfg RedisConfig
var test_page_size = 5
var test_key = "gotest"

func init() {
   
}


func TestPageCache(t *testing.T) {
    cache, err := NewPageCache(&pct_cfg, test_page_size)
    if err != nil {
        t.Fatal(err)
    }
    _, err = cache.DelAll(test_key)
    if err != nil {
        t.Fatal(err)
    }

    for i:=int64(1);i<17; i++ {
        err = cache.Add(test_key, i)
        if err != nil {
            t.Fatal(err)
        }
    }

    var page int
    for i:=int64(3);i<16;i+=2 {
        page = int(math.Ceil(float64(i)/5.0))
        err = cache.Del(test_key, page, i)
        if err != nil {
            t.Fatal(err)
        }
    }
    for i:=int64(11);i<16;i++ {
        page = int(math.Ceil(float64(i)/5.0))
        err = cache.Del(test_key, page, i)
        if err != nil {
            t.Fatal(err)
        }
    }

}

var str = `
for i=11,15 do
    local page = math.ceil(i/5.0)
    local ok, err = cache:del(test_key, page, i)
end
`