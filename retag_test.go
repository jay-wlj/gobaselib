package base

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

type Product struct {
	UserId     int64     `json:"user_id,omitempty"`
	Title      string    `json:"title"`
	Status     int16     `json:"status,omitempty"`
	CreateTime time.Time `json:"create_time"`
	SkuName    string    `json:"-"`
}

func TestRetag(t *testing.T) {
	v := Product{UserId: 1001, Title: "苹果", Status: 1, CreateTime: time.Now(), SkuName: "512G"}
	body, _ := json.Marshal(FilterStruct(&v, false, "user_id", "status"))
	t.Log("body:", string(body))

	body, _ = json.Marshal(FilterStruct(&v, true, "user_id", "status"))
	t.Log("body:", string(body))
}

func TestRemoveUint64Slice(t *testing.T) {
	fmt.Println(RemoveUint64Slice([]uint64{1, 2, 3, 4, 10}, []uint64{1, 2, 4, 10}))
}
