package db

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// 公共model更新时间及创建时间
type Model struct {
	Id         int64 `json:"id" gorm:"primary_key:id"`
	CreateTime int64 `json:"create_time"`
	UpdateTime int64 `json:"update_time"`
}

func (t *Model) BeforeCreate(scop *gorm.Scope) error {
	fmt.Println("BeforeCreate")
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
	fmt.Println("BeforeUpdate")
	scop.SetColumn("update_time", time.Now().Unix())
	return nil
}
