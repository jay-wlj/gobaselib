package cache_redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jay-wlj/gobaselib/log"
	"gorm.io/driver/mysql"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dsn    = "root:123456@tcp(10.10.21.42:3306)/config_server?charset=utf8&parseTime=true&interpolateParams=true"
	db     *gorm.DB
	client *redisCache
)

func init() {
	MakeCacheQuery(&UserCacheQuery)

	var err error
	client, err = NewClient(&Config{
		Network:  "tcp",
		Nodes:    []string{"10.10.21.41:6379"},
		Password: "meiliredis123",
	})
	if err != nil {
		log.Error("NewClient fail! err=", err)
		return
	}
	db, _ = gorm.Open(mysql.Open(dsn))
}

type Member struct {
	Id     uint64 `json:"id"`
	Accunt string `json:"accunt"`
	Name   string `json:"name"`
}

func (t *Member) TableName() string {
	return "member"
}

var UserCacheQuery func(context.Context, *redisCache, string, time.Duration, func(int64, int64) ([]Member, error), int64, int64) ([]Member, error, string)

func getUserPage(start, end int64) (vs []Member, err error) {

	if err = db.Model(&Member{}).Order("id asc").Offset(int(start)).Limit(int(end - start)).Find(&vs).Error; err != nil {
		log.Error("getUserPage fail! err=", err)
		return
	}
	return
}

func TestUserCache(t *testing.T) {
	ctx := context.TODO()
	fmt.Println(UserCacheQuery(ctx, client, "abc-1-10", 3*time.Second, getUserPage, 1, 10))
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))
	time.Sleep(time.Second)
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))
	time.Sleep(time.Second)
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))
	time.Sleep(time.Second)
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))
	time.Sleep(time.Second)
	fmt.Println(UserCacheQuery(ctx, client, "abc-10-20", 3*time.Second, getUserPage, 10, 20))

	fmt.Println(client.Exists(ctx, "def").Result())
}
