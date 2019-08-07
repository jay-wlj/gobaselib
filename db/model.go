package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

// 公共model更新时间及创建时间
type Model struct {
	Id         int64 `json:"id" gorm:"primary_key:id" view:"*"`
	CreateTime int64 `json:"create_time,omitempty"`
	UpdateTime int64 `json:"update_time,omitempty"`
}

func (t *Model) BeforeCreate(scop *gorm.Scope) error {
	//fmt.Println("BeforeCreate")
	scop.SetColumn("create_time", time.Now().Unix())
	scop.SetColumn("update_time", time.Now().Unix())
	return nil
}

// func (t *Model) BeforeSave(scop *gorm.Scope) error {
// 	fmt.Println("BeforeSave")
// 	scop.SetColumn("update_time", time.Now().Unix())
// 	return nil
// }

func (t *Model) BeforeUpdate(scop *gorm.Scope) error {
	//fmt.Println("BeforeUpdate")
	scop.SetColumn("update_time", time.Now().Unix())
	// if scop.HasColumn("create_time") {
	// 	scop.SetColumn("create_time", gorm.Expr("create_time"))
	// }
	scop.Search.Omit("create_time") // 忽略更新create_time字段
	return nil
}
