package cache

import (
    "testing"
    "fmt"
)


func TestSlice(t *testing.T) {
    var count = 3
    var arr = make([]int, count)
    for i:=0;i<count -1 ;i++ {
        arr[i] = i
    }
    fmt.Printf(">> %v\n", arr)
}

func cache_get(key string)(data string, err error) {
    return "", fmt.Errorf("invalid")
}

func get_last_page(key string)(err error){
    // 获取当前页(TODO: 添加缓存)
    data, err := cache_get(key)
    if err != nil {
        return 
    }
    fmt.Printf("data: %v\n", data)
    
    return
}

func TestError(t *testing.T){
    err := get_last_page("xxxx")
    t.Logf("err: %v", err)
}

// func TestSliceFor(t *testing.T){
//     next_page := 0
//     for i:=0;i< 5;i++ {
//         t.Logf("000 next_page: %d", next_page)
//         data, next_page := GetData(next_page)
//         t.Logf("data: %v", data)
//         t.Logf("111 next_page: %d", next_page)
//     }
// }