package db

import (
	"fmt"
	"time"

	"github.com/jie123108/glog"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type MongoSessionDao struct {
	mongo_url string
	timeout   time.Duration
	session   *mgo.Session
}

type MongoDao struct {
	coll    *mgo.Collection
	session *mgo.Session
}

var g_mongosessionmap map[string]*MongoSessionDao

func InitMgo(mongo_url string, timeout time.Duration) (*mgo.Session, error) {
	var err error
	newsession, err := mgo.DialWithTimeout(mongo_url, timeout)
	if err != nil {
		glog.Errorf("open mongodb(%s) failed! err: %v", mongo_url, err)
		return nil, err
	} else {
		glog.Infof("open mongodb(%s) ok!", mongo_url)
	}
	newsession.SetMode(mgo.Eventual, true)
	return newsession, nil
}

func NewMongoDao(mongo_url string, timeout time.Duration, collection string) (*MongoDao, error) {
	if nil == g_mongosessionmap {
		g_mongosessionmap = make(map[string]*MongoSessionDao)
	}

	session := g_mongosessionmap[mongo_url]

	if nil == session {
		session = &MongoSessionDao{mongo_url, timeout, nil}
		var err error
		session.session, err = InitMgo(mongo_url, timeout)
		if err != nil {
			return nil, err
		}
		session.session.SetPoolLimit(1024)
		//session.session.SetMode(mgo.Monotonic, true)
		g_mongosessionmap[mongo_url] = session
	}

	dao := &MongoDao{nil, session.session.Clone()}
	dao.coll = session.session.DB("").C(collection)

	if err := dao.session.Ping(); err != nil {
		// TODO Refersh 失败, 要处理
		session.session.Refresh()
	}
	return dao, nil
}

func (this *MongoDao) Close() {
	if this == nil {
		return
	}
	if this.session != nil {
		this.session.Close()
		this.session = nil
	}
}

func (this *MongoDao) getAutoIdCollection() *mgo.Collection {
	var coll *mgo.Collection
	//coll = autoidcoll[this.mongo_url]
	//if coll == nil {
	coll = this.session.DB("").C("autoid")
	//autoidcoll[this.mongo_url] = coll
	//}
	return coll
}

type AutoId struct {
	Name   string `json:"name" bson:"name"`
	Autoid int    `json:"autoid" bson:"autoid"`
}

func (this *MongoDao) GetNextId(key string) (newId int, err error) {
	coll := this.getAutoIdCollection()
	if coll == nil {
		return 0, fmt.Errorf("getAutoIdCollection failed")
	}

	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"autoid": 1}},
		Upsert:    true,
		ReturnNew: true,
	}

	var doc = AutoId{}
	info, err := coll.Find(bson.M{"name": key}).Apply(change, &doc)
	glog.Infof("coll.Find.Apply(%v)", info)
	newId = doc.Autoid

	return
}

func (this *MongoDao) EnsureIndexs(keys []string) (err error) {
	for _, key := range keys {
		index := mgo.Index{
			Key:        []string{key},
			Unique:     false,
			DropDups:   false,
			Background: true,
			Sparse:     true,
		}
		err = this.coll.EnsureIndex(index)
		if err != nil {
			glog.Errorf("EnsureIndexs key:%s,error:%s", key, err.Error())
			return
		}
	}
	return
}

func (this *MongoDao) Insert(value interface{}) (err error) {
	return this.coll.Insert(value)
}

func (this *MongoDao) Update(selector interface{}, value interface{}) (err error) {
	return this.coll.Update(selector, value)
}

func (this *MongoDao) UpdateAll(selector interface{}, value interface{}) (info *mgo.ChangeInfo, err error) {
	return this.coll.UpdateAll(selector, value)
}

func (this *MongoDao) Upsert(selector interface{}, value interface{}) error {
	info, err := this.coll.Upsert(selector, value)
	glog.Info("coll.upsert(%v), err:%v", info, err)

	return err
}

func (this *MongoDao) Delete(selector interface{}) (err error) {
	return this.coll.Remove(selector)
}
func (this *MongoDao) Find(selector interface{}, fields interface{}, page int, page_size int, result interface{}, sortfields ...string) (err error) {
	query := this.coll.Find(selector).Sort(sortfields...)

	skip := 0
	if page > 1 {
		skip = (page - 1) * page_size
	}
	if skip > 0 {
		query = query.Skip(skip)
	}

	if page_size > 0 {
		query = query.Limit(page_size)
	}
	query = query.Select(fields)

	err = query.All(result)
	if err != nil {
		glog.Errorf("err:%v query:%v", err, query)
	}
	return
}

func (this *MongoDao) Count(selector interface{}) (int, error) {
	query := this.coll.Find(selector)
	count, err := query.Count()
	return count, err
}

func (this *MongoDao) One(selector interface{}, fields interface{}, result interface{}) (err error) {
	query := this.coll.Find(selector)

	selected_query := query.Select(fields)

	err = selected_query.One(result)
	if err == mgo.ErrNotFound {
		err = nil
	}
	return
}

func (this *MongoDao) FindOne(selector interface{}, fields interface{}, result interface{}, sortfields ...string) (err error) {
	query := this.coll.Find(selector).Sort(sortfields...)

	selected_query := query.Select(fields)

	err = selected_query.One(result)
	if err == mgo.ErrNotFound {
		err = nil
	}
	return
}

func (this *MongoDao) Pipe(pipeline interface{}, result interface{}) (err error) {
	err = this.coll.Pipe(pipeline).All(result)
	return
}
