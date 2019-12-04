package push

import (
	//"bytes"
	//"fmt"
	//"github.com/fatih/structs"
	//"github.com/jie123108/gocrc48"
	//"github.com/jie123108/glog"
	//base "github.com/jay-wlj/gobaselib"
	"github.com/jay-wlj/gobaselib/db"
	"gopkg.in/mgo.v2/bson"
	//"time"
)

type JPushBindDao struct {
	dao *db.MongoDao
}

func NewJPushBindDao() (*JPushBindDao, error) {
	jpushbinddao := &JPushBindDao{}
	mongo_url, timeout := pushconfig.MongoUrl, pushconfig.TimeOut

	var err error
	jpushbinddao.dao, err = db.NewMongoDao(mongo_url, timeout, "jpushbind")

	return jpushbinddao, err
}

func (this *JPushBindDao) Close() {
	this.dao.Close()
}

type ValueInfo struct {
	RegId  string `json:"reg_id" bson:"reg_id"`
	UserId int64  `json:"user_id" bson:"user_id"`
	AppKey string `json:"appkey" bson:"appkey"`
}

func (this *JPushBindDao) Upsert(reg_id string, appkey string, user_id int64) (err error) {
	selector := bson.M{"reg_id": reg_id}
	values := ValueInfo{reg_id, user_id, appkey}
	err = this.dao.Upsert(selector, values)
	return
}

func (this *JPushBindDao) FindByUserIds(user_ids []int64) ([]ValueInfo, error) {
	selector := bson.M{"user_id": bson.M{"$in": user_ids}}
	values := []ValueInfo{}
	fields := bson.M{}
	err := this.dao.Find(selector, fields, 0, 0, &values)
	return values, err
}

func (this *JPushBindDao) FindByUserId(user_id int64) (ValueInfo, error) {
	selector := bson.M{"user_id": user_id}
	value := ValueInfo{}
	fields := bson.M{}
	err := this.dao.One(selector, fields, &value)
	return value, err
}

func (this *JPushBindDao) FindByRegId(reg_id string) (ValueInfo, error) {
	selector := bson.M{"reg_id": reg_id}
	value := ValueInfo{}
	fields := bson.M{}
	err := this.dao.One(selector, fields, &value)
	return value, err
}

func (this *JPushBindDao) DeleteByRegId(reg_id string) (err error) {
	selector := bson.M{"reg_id": reg_id}
	err = this.dao.Delete(selector)
	return
}
