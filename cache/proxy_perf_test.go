package cache

import (
    "testing"    
    "reflect"
    "fmt"
    "time"
)


func PerfCacheQuery(in []reflect.Value) []reflect.Value {
    query_func := in[0]
    args := in[1:]
    values := query_func.Call(args)
    return values
}

func MakePerfCacheQuery(fptr interface{}){
    fn := reflect.ValueOf(fptr).Elem()
    v := reflect.MakeFunc(fn.Type(), PerfCacheQuery)
    fn.Set(v)
}

type TestObject struct {
    UserId int64
    Username string
    Sex bool
    Age int 
    Avatar string
}

func GetById(obj_id int64) (obj *TestObject, err error){
    // objname := fmt.Sprintf("obj-%d", obj_id)
    // sex := obj_id % 2 == 0
    // age := int(obj_id % 60)
    // avatar := "avagar" 
    // obj = &TestObject{UserId: obj_id, Username: objname,
    //         Sex: sex, Age: age, Avatar: avatar}

    obj = &TestObject{}
    return
}

var ObjectCacheQuery func (
    func(int64) (*TestObject, error), int64) (*TestObject, error)

func init(){
    MakePerfCacheQuery(&ObjectCacheQuery)
}

func TestPerf(t *testing.T){
    // var obj *TestObject 
    // var err error
    count := 10000*10
    obj_id := int64(100)

    start := time.Now()
    for i:=0;i<count;i++ {
        ObjectCacheQuery(GetById, obj_id)
    }
    end := time.Now()
    fmt.Printf("proxy used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)

    start = time.Now()
    for i:=0;i<count;i++ {
        GetById(obj_id)
    }
    end = time.Now()
    fmt.Printf("raw used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)
}