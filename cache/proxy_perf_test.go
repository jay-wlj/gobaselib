package cache

import (
    "testing"    
    "reflect"
    "fmt"
    "time"
)


func perfCacheQuery(in []reflect.Value) []reflect.Value {
    query_func := in[0]
    args := in[1:]
    values := query_func.Call(args)
    return values
}

func makePerfCacheQuery(fptr interface{}){
    fn := reflect.ValueOf(fptr).Elem()
    v := reflect.MakeFunc(fn.Type(), perfCacheQuery)
    fn.Set(v)
}

type testObject struct {
    UserId int64
    Username string
    Sex bool
    Age int 
    Avatar string
}

func getById(obj_id int64) (obj *testObject, err error){
    // objname := fmt.Sprintf("obj-%d", obj_id)
    // sex := obj_id % 2 == 0
    // age := int(obj_id % 60)
    // avatar := "avagar" 
    // obj = &TestObject{UserId: obj_id, Username: objname,
    //         Sex: sex, Age: age, Avatar: avatar}

    obj = &testObject{}
    return
}

var ObjectCacheQuery func (
    func(int64) (*testObject, error), int64) (*testObject, error)

func init(){
    //makePerfCacheQuery(&ObjectCacheQuery)
}

func testPerf(t *testing.T){
    // var obj *TestObject 
    // var err error
    count := 10000*10
    obj_id := int64(100)

    start := time.Now()
    for i:=0;i<count;i++ {
        ObjectCacheQuery(getById, obj_id)
    }
    end := time.Now()
    fmt.Printf("proxy used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)

    start = time.Now()
    for i:=0;i<count;i++ {
        getById(obj_id)
    }
    end = time.Now()
    fmt.Printf("raw used: %.3f\n", float64(end.Sub(start).Nanoseconds())/1000000000.0)
}