package db

import (
	"log"
	"os"
	"strings"

	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
)

func init_db(sql_url string, sql_debug bool) (db *gorm.DB) {
	var err error
	db, err = gorm.Open("postgres", sql_url)
	if err != nil {
		glog.Exitf("open psql(%s) failed! err: %v", sql_url, err)
	} else {
		glog.Infof("open psql(%s) ok!", sql_url)
	}
	// TODO: 数据库连接池相关参数设置。
	// DB.SetConnMaxLifetime(d)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(100)

	db.LogMode(sql_debug)

	if sql_debug {
		db.SetLogger(log.New(os.Stdout, "upload:", log.Lmicroseconds))
	}

	return
}

var m_psqlDb map[string]*gorm.DB = make(map[string]*gorm.DB)

func InitPsqlDb(psqlUrl string, sql_debug bool) (db *gorm.DB) {
	var ok bool
	if db, ok = m_psqlDb[psqlUrl]; ok {
		return
	}

	db = init_db(psqlUrl, sql_debug)

	m_psqlDb[psqlUrl] = db
	glog.Infof("open psql(%s) ok!", psqlUrl)
	return
}

type PsqlDao struct {
	model    interface{}
	psqlUrl  string
	sqlDebug bool
	db       *gorm.DB
	useTx    bool
}

func NewPsqlDao(psqlUrl string, sqlDebug bool, model interface{}) (dao *PsqlDao) {
	dao = &PsqlDao{model: model, psqlUrl: psqlUrl, sqlDebug: sqlDebug, db: InitPsqlDb(psqlUrl, sqlDebug)}
	return
}

func (this *PsqlDao) Insert(value interface{}) (err error) {
	db := this.db.Create(value)
	err = db.Error
	return
}

func (this *PsqlDao) Upsert(obj interface{}) (err error, operator string) {
	operator = "save"
	err = this.db.Create(obj).Error
	if err != nil && strings.Index(err.Error(), "#1062") > 0 {
		operator = "update"
		err = this.db.Model(this.model).Update(obj).Error
	}

	return
}

func (this *PsqlDao) Update(update interface{}, query interface{}, args ...interface{}) (err error) {
	err = this.db.Model(this.model).Where(query, args).Update(update).Error
	return
}

func (this *PsqlDao) Find(result interface{}, query interface{}, args ...interface{}) (err error) {
	err = this.db.Where(query, args).Find(result).Error
	return
}

func (this *PsqlDao) Find2(result interface{}, order_by interface{}, page int, page_size int, query interface{}, args ...interface{}) (err error) {
	if page > 0 {
		page = page - 1
	}
	if page_size < 1 {
		page_size = 1
	}
	offset := page * page_size
	limit := page_size
	err = this.db.Model(this.model).Where(query, args).Order(order_by).Offset(offset).Limit(limit).Find(result).Error
	return
}

func (this *PsqlDao) Get(result interface{}, query interface{}, args ...interface{}) (err error) {
	err = this.db.Where(query, args).First(result).Error
	return
}

func (this *PsqlDao) Exec(sql string, values ...interface{}) (affected_rows int64, err error) {
	db := this.db.Exec(sql, values)
	affected_rows = db.RowsAffected
	err = db.Error
	return
}
