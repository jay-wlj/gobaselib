package cache

import (
    "testing"
    "fmt"
    "time"
    "encoding/json"
    "github.com/vmihailenco/msgpack"
)
type Comment struct {
    Id             int64       `json:"id,string" msgpack:"id"`
    UserId         int64       `json:"user_id,string" msgpack:"user_id"`
    Type           int         `json:"type" msgpack:"type"`
    ResId          int64       `json:"res_id,string" msgpack:"res_id"`
    Comment        string      `json:"comment" msgpack:"comment"`
}

func TestMsgPack(t *testing.T) {
    count := 10000 * 10
    Id := int64(100)
    type_ := 1
    ResId := int64(10)
    UserId := int64(100)
    comment := "comment"
    commentInfo := &Comment{Id,UserId, type_, ResId, comment}
    commentBuf,_ := json.Marshal(commentInfo)
    commentPack,_ := msgpack.Marshal(commentInfo)
    // fmt.Printf("%v\n%v\n", commentBuf, commentPack)
    start := time.Now()
    for i:=0;i<count;i++ {
        json.Marshal(commentInfo)
    }
    end := time.Now()
    fmt.Printf("json marshal used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)
   
    start = time.Now()
    for i:=0;i<count;i++ {
        json.Unmarshal(commentBuf, commentInfo)
    }
    end = time.Now()
    fmt.Printf("json unmarshal used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)

    start = time.Now()
    for i:=0;i<count;i++ {
        msgpack.Marshal(commentInfo)
    }
    end = time.Now()
    fmt.Printf("msgpack marshal used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)

    start = time.Now()
    for i:=0;i<count;i++ {
        msgpack.Unmarshal(commentPack)
    }
    end = time.Now()
    fmt.Printf("msgpack unmarshal used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)


}
