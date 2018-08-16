package db

import (
	"database/sql"

	"github.com/jie123108/glog"
	// "github.com/ziutek/mymysql/godrv"
	//"encoding/json"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/gorp.v1"
	//"strconv"
	"strings"
)

type MysqlDao struct {
	mysql_url string
	sql_debug bool
	dbmap     *gorp.DbMap
	tx        *gorp.Transaction
	usertx    bool
}

var g_db *sql.DB

func InitDb(mysql_url string) (*sql.DB, error) {
	if g_db != nil {
		glog.Infof("InitDb(mysql_url string) g_db already exists")
		return g_db, nil
	}

	var err error
	db, err := sql.Open("mysql", mysql_url)
	if err != nil {
		glog.Fatalf("open mysql(%s) failed! err:%v", mysql_url, err)
		return nil, err
	} else {
		glog.Infof("open mysql(%s) ok!", mysql_url)
	}
	g_db = db

	//TODO: 数据库连接池相关参数设置。
	// g_db.SetConnMaxLifttime(d)
	g_db.SetMaxIdleConns(500)
	g_db.SetMaxOpenConns(2000)

	return g_db, err
}

func NewMysqlDao(mysql_url string, sql_debug bool, tablename string, tablestu interface{}, isautoid bool, keys []string) (dao *MysqlDao, err error) {
	dao = &MysqlDao{mysql_url, sql_debug, nil, nil, false}
	db, err := InitDb(mysql_url)
	if err != nil {
		return nil, err
	}

	if dao.dbmap == nil {
		// construct a gorp DbMap
		dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "utf8mb4"}}
		if sql_debug {
			var traceon = "["
			traceon += tablename
			traceon += "dao]"
			dbmap.TraceOn(traceon, log.New(os.Stdout, "upload:", log.Lmicroseconds))
		}
		dao.dbmap = dbmap
	}

	if len(tablename) > 0 {
		dao.dbmap.AddTableWithName(tablestu, tablename).SetKeys(isautoid, keys...)
	}
	return
}

func (this *MysqlDao) AddTableWithName(tablestu interface{}, tablename string, isautoid bool, keys []string) {
	glog.Errorf("addtablewithname(%v)", tablename)
	this.dbmap.AddTableWithName(tablestu, tablename).SetKeys(isautoid, keys...)
	return
}

func (this *MysqlDao) Begin() (err error) {
	this.tx, err = this.dbmap.Begin()
	this.usertx = true
	return
}

func (this *MysqlDao) Commit() (err error) {
	if this.usertx && this.tx != nil {
		err = this.tx.Commit()
	}
	return
}

func (this *MysqlDao) Rollback() (err error) {
	if this.usertx && this.tx != nil {
		err = this.tx.Rollback()
	}
	return
}

func (this *MysqlDao) Insert(obj interface{}) (err error) {
	if this.usertx {
		if this.tx == nil {
			return fmt.Errorf("this.dbmap is nil")
		}
		err = this.tx.Insert(obj)
	} else {
		if this.dbmap == nil {
			return fmt.Errorf("this.dbmap is nil")
		}
		err = this.dbmap.Insert(obj)
	}
	return
}

func (this *MysqlDao) Upsert(obj interface{}) (err error, effects int64, operator string) {
	if this.usertx {
		operator = "save"
		if this.tx == nil {
			return fmt.Errorf("this.dbmap is nil"), effects, operator
		}
		err = this.tx.Insert(obj)
		if err != nil && strings.Index(err.Error(), "#1062") > 0 {
			operator = "update"
			effects, err = this.tx.Update(obj)
		}
	} else {
		operator = "save"
		if this.dbmap == nil {
			return fmt.Errorf("this.dbmap is nil"), effects, operator
		}
		err = this.dbmap.Insert(obj)
		if err != nil && strings.Index(err.Error(), "#1062") > 0 {
			operator = "update"
			effects, err = this.dbmap.Update(obj)
		}
	}
	return err, effects, operator
}

func (this *MysqlDao) Update(obj interface{}) (err error, rows int64) {
	if this.usertx {
		if this.tx == nil {
			return fmt.Errorf("this.db is nil"), 0
		}
		rows, err = this.tx.Update(obj)
	} else {
		if this.dbmap == nil {
			return fmt.Errorf("this.db is nil"), 0
		}
		rows, err = this.dbmap.Update(obj)
	}
	return err, rows
}

func (this *MysqlDao) Exec(sqlstr string, args ...interface{}) (sql.Result, error) {
	if this.usertx {
		return this.tx.Exec(sqlstr, args...)
	} else {
		return this.dbmap.Exec(sqlstr, args...)
	}
}

func (this *MysqlDao) Select(results interface{}, sql_select string, page int, page_size int, args ...interface{}) ([]interface{}, error) {
	offset := 0
	if page > 0 {
		offset = (page - 1) * page_size
	}
	if page_size > 0 {
		sql_select += fmt.Sprintf(" limit %v,%v", offset, page_size)
	}

	if this.usertx {
		return this.tx.Select(results, sql_select, args...)
	} else {
		return this.dbmap.Select(results, sql_select, args...)
	}
}

type CountStruct struct {
	Count int `json:"count" db:"count"`
}

func (this *MysqlDao) Count(sql_count string) (int, error) {
	total := 0
	var err error
	results := []CountStruct{}
	_, err = this.Select(&results, sql_count, 0, 0)

	if err == nil {
		if len(results) > 0 {
			total = results[0].Count
		}
	}

	return total, err
}
