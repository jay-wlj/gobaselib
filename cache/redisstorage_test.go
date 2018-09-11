package cache

import (
    "testing"
    "fmt"
    "reflect"
    // "encoding/json"
    "github.com/vmihailenco/msgpack"
)


type comment struct {
    Id             int64       `json:"id,string" db:"id"`
    UserId         int64       `json:"user_id,string" db:"user_id"`
    Type           int         `json:"type" db:"type"`
    ResId          int64       `json:"res_id,string" db:"res_id"`
    Comment        string      `json:"comment" db:"comment"`
}

func commentGetID(obj interface{}) int64 {
    comment := obj.(*comment)
    return comment.Id
}

func commentGetKey(id int64) string {
    return fmt.Sprintf("cmt:%d", id)
}

func indexKey(type_ int, res_id int64) string {
    return fmt.Sprintf("cidx:%d-%d", type_, res_id)
}

func commentGetIndexKey(obj interface{}) string {
    comment := obj.(*comment)
    return indexKey(comment.Type, comment.ResId)
}

func commentMarshal(obj interface{})([]byte, error) {
    return msgpack.Marshal(obj)
}

func commentUnmarshal(data []byte, v interface{}) error {
    return msgpack.Unmarshal(data, v)
}


func getByPage(comments []*comment, page int) (objs []*comment, err error) {
    if page == 0 {
        page = (len(comments)-1)/10 + 1
    }
    // fmt.Printf("page: %d\n", page)
    b := (page-1) * 10
    e := page * 10
    if e > len(comments) {
        e = len(comments)
    }
    objs = comments[b:e]
    tmp := make([]*comment, 0)
    //倒序插入.
    for i := len(objs)-1;i >=0; i-- {
        obj := objs[i]
        if obj.Id != -1 { //-1表示标记为删除
            tmp = append(tmp, obj)
        }
    }
    objs = tmp
    // 本页已经全部完毕, 取下一页.
    if len(objs) == 0 {
        page = page -1
        if page > 0 {
            return getByPage(comments, page)
        }
    }

    return objs, nil
}

func getTotal(comments []*comment)(total int64) {
    for i := len(comments)-1;i >=0; i-- {
        obj := comments[i]
        if obj.Id != -1 { //-1表示标记为删除
            total += 1
        }
    }
    return
}

func commentsCompare(comments []interface{}, comments_expect []*comment) (err error) {
    exp_len := len(comments_expect)
    real_len := len(comments)
    if real_len !=  exp_len{
        err = fmt.Errorf("expect %d comments, but got %d", exp_len, real_len)
        return
    }
    for i:=0;i<real_len;i++ {
        comment := comments[i].(*Comment)
        comment_exp := comments_expect[i]
        if comment.Id != comment_exp.Id {
            err = fmt.Errorf("idx [%d], expect comment: %d, but got: %d", i, comment.Id, comment_exp.Id)
            return
        }else if comment.Comment != comment_exp.Comment {
            err = fmt.Errorf("index [%d], expect comment: %d:%s, but got: %d:%s",
                    i, comment.Id, comment.Comment, comment_exp.Id, comment_exp.Comment)
        }
    }
    return
}

func test_comments_del(comments []*comment, id int64) {
    for i:=0;i<len(comments); i++ {
        if comments[i].Id == id {
            comments[i].Id = -1 //标记为-1表示删除
            break
        }
    }
}

func test_redis_list(t *testing.T, storage *RedisStorage, indexkey string, comments_all []*comment) {
    next_page := 0
    total_objs := 0
    for i:=0;i<len(comments_all)/10 + 3;i++ {
        cur_page := next_page
        objs_expect, _ := getByPage(comments_all, cur_page)
        data, err := storage.List(indexkey, cur_page)
        if err != nil {
            t.Fatal(err)
        }
        total_exp := getTotal(comments_all)
        if data.Total != total_exp {
            t.Fatalf("total comments: %d, expect totals: %d", data.Total, total_exp)
        }
        next_page = data.NextPage
        total_objs += len(data.Objs)
        err = commentsCompare(data.Objs, objs_expect)
        if err != nil {
            t.Errorf("    objs: %v", data.Objs)
            t.Errorf("exp objs: %v", objs_expect)    
            t.Fatal(fmt.Errorf("check page [%s:%d] %v", indexkey, cur_page, err))
        }else{
            t.Logf("check page [%s:%d] success!", indexkey, cur_page)
        }
        if next_page < 0 {
            break
        }
    }
    t.Logf("--------- check list [%d] ---------", total_objs)
    
}


func testRedisStorage(t *testing.T) {
    cfg := &StorageConfig{}
    objtype := reflect.TypeOf((*comment)(nil)).Elem()
    cb := &StorageCallback{commentGetID, commentGetKey, commentGetIndexKey, commentMarshal, 
                commentUnmarshal, objtype}
    storage, err := NewRedisStorage(cfg, cb)
    if err != nil {
        t.Fatal(err)
    }

    type_ := 1
    ResId := int64(10)
    indexkey := indexKey(type_, ResId)
    _, err = storage.DelAllIndex(indexkey)
    if err != nil {
        t.Fatal(err)
    }

    comments := make([]*comment, 0)
    del_ids := make([]int64, 0)

    for i:=int64(1);i<46;i++ {
        Id := i
        UserId := i * 3
        comment := fmt.Sprintf("comment-%d-%d: %d", type_, ResId, Id)
        commentInfo := &comment{Id,UserId, type_, ResId, comment}
        err = storage.Add(commentInfo)
        if err != nil {
            t.Fatal(err)
        }
        comment_update := fmt.Sprintf("comment-update-%d-%d: %d", type_, ResId, Id)
        commentInfo.Comment = comment_update
        err = storage.Update(commentInfo)
        if err != nil {
            t.Fatal(err)
        }
        
        comments = append(comments, commentInfo)
        if i % 5 == 0 {
            t.Logf("add %d comments", i)
            test_redis_list(t, storage, indexkey, comments)
        }
        if i % 3 == 0 || i < 11 || i > 35 {
            del_ids = append(del_ids, Id)
        }
    }


    for _, Id := range(del_ids) {
        t.Logf("Del Comment: %d", Id)
        _, err = storage.Del(indexkey, Id)
        if err != nil {
            t.Fatal(err)
        }
        test_comments_del(comments, Id)
        test_redis_list(t, storage, indexkey, comments)    
    }
    test_redis_list(t, storage, indexkey, comments)

}
