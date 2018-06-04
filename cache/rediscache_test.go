package cache

import (
    "testing"
    "fmt"
    "time"
    // "encoding/json"
)

var (
    cfg RedisConfig
)


type User struct {
    UserId int64
    Username string
    Sex bool
    Age int 
    Avatar string
}

func GetUserById(user_id int64) (user *User, err error){
    username := fmt.Sprintf("user-z-%d", user_id)
    sex := user_id % 2 == 0
    age := int(user_id % 60)
    avatar := fmt.Sprintf("http://img.kc.cn/path/to/z/%d.jpg", user_id)
    user = &User{UserId: user_id, Username: username,
            Sex: sex, Age: age, Avatar: avatar}
    return
}

func GetUser(user_id int64, username string, sex bool, age int, avatar string) (user *User, err error){
    user = &User{UserId: user_id, Username: username,
            Sex: sex, Age: age, Avatar: avatar}
    return
}

type Object struct {
    name string
}

func (this *Object) GetEmptyUser() (user *User, err error){
    user = &User{}
    user.Username = this.name
    return
}

var UserCacheQuery func (
    *RedisCache, string, time.Duration, func(int64) (*User, error), int64) (*User, error, string)
var UserExCacheQuery func (
    *RedisCache, string, time.Duration, func(int64, string, bool, int, string) (*User, error), int64, string, bool, int, string) (*User, error, string)
var EmptyUserCacheQuery func (
    *RedisCache, string, time.Duration, func() (*User, error)) (*User, error, string)

var article_id int64

type Article struct {
    UserId int64
    Id int64
    Type_ int 
    Title string
    Content string
}
var GetArticlesCacheQuery func (
    *RedisCache, string, time.Duration, func(int64, int, int) ([]Article, error),
     int64, int, int) ([]Article, error, string)


func init(){
    MakeCacheQuery(&UserCacheQuery)
    MakeCacheQuery(&UserExCacheQuery)
    MakeCacheQuery(&EmptyUserCacheQuery)
    article_id = 1
    MakeCacheQuery(&GetArticlesCacheQuery)
}



func GetArticles(user_id int64, type_ int, page int) (articles []Article, err error){
    fmt.Printf("Get Articles from database ...\n")
    articles = make([]Article, 3)
    for i := 0;i<3;i++ {
        articles[i].UserId = user_id
        articles[i].Id = article_id 
        articles[i].Type_ = type_
        articles[i].Title = fmt.Sprintf("title: %v", article_id)
        articles[i].Content = fmt.Sprintf("content: %v", article_id)
        article_id += 1
    }
    return
}

func TestCache(t *testing.T) {
    cache, err := NewRedisCache(&cfg)
    if err != nil {
        t.Fatal(err)
    }
    val, err := cache.Get("not-exist-key")
    if err == ErrNotExist {
        t.Logf("not exist")
    }else if err != nil {
        t.Errorf("val:%s, err: [%v]", val, err)
        return
    }

    key := "unit-001"
    err = cache.Set(key, 3, time.Hour)
    if err != nil {
        t.Errorf("cache set failed! err: %v", err)
        return
    }

    val, err = cache.Get(key)
    if err != nil {
        t.Errorf("cache get failed! err: %v", err)
        return
    }

    n, err := cache.Del(key)
    if err != nil {
        t.Errorf("cache del failed! err: %v", err)
        return
    }

    t.Logf("del n: %v, err: %v", n, err)
}

func TestCacheQuery(t *testing.T){
    cache, err := NewRedisCache(&cfg)
    if err != nil {
        t.Fatal(err)
    }

    user_id := int64(10)
    cachekey := fmt.Sprintf("u:%d", user_id)
    
    user, err, cached := UserCacheQuery(cache, cachekey, time.Hour, GetUserById, user_id)
    t.Logf("val: %v, err:%v, cached: %v", user, err, cached)

    user_id = int64(20)
    cachekey = fmt.Sprintf("u:%d", user_id)
    user, err, cached = UserExCacheQuery(cache, cachekey, time.Hour, GetUser, 
                    user_id, "lxj", true, 33, "avatar.jpg")
    t.Logf("val: %v, err:%v, cached: %v", user, err, cached)

    user_id = int64(30)
    cachekey = fmt.Sprintf("u:%d", user_id)
    obj := Object{"empty obj"}
    user, err, cached = EmptyUserCacheQuery(cache, cachekey, time.Hour, obj.GetEmptyUser)
    t.Logf("val: %v, err:%v, cached: %v", user, err, cached)
}

func TestCacheQueryEx(t *testing.T){

    cache, err := NewRedisCache(&cfg)
    if err != nil {
        t.Fatal(err)
    }
    user_id := int64(100)
    // page := 1
    // type_ := 2
    // articles, _ := GetArticles(user_id, type_, page)
    // str, _ := json.Marshal(articles)
    // fmt.Printf("-----:::: %v", string(str))
    key := fmt.Sprintf("arts:%v", user_id)
    for type_ := 3;type_ <= 5; type_+=2 {
        for page := 1; page < 3; page++ {
            field := fmt.Sprintf("%d-p%d", type_, page)
            cachekey := fmt.Sprintf("%s->%s", key, field)
            articles, _, cached := GetArticlesCacheQuery(cache, cachekey, time.Hour, 
                GetArticles, user_id, type_, page)
            t.Logf("%v  articles: %v", cached, articles)
        }
    }
    
}