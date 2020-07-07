package db

import (
	"errors"
	"fmt"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	"github.com/jie123108/glog"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var m_psqlDb map[string]*gorm.DB

func init() {
	m_psqlDb = make(map[string]*gorm.DB)
}

// 默认db连接
var m_db *gorm.DB

type PsqlDB struct {
	*gorm.DB
	tx_flag         int32    // 事务标志
	fn_after_commit []func() // 提交后执行的代码
}

// func (v *PsqlDB) GetDB() *gorm.DB {
// 	return v.DB
// }

func (v *PsqlDB) ListPage(page, page_size int) *gorm.DB {
	if page <= 0 {
		page = 1
		glog.Error("ListPage page:", page, " <= 0")
	}
	if page_size <= 0 {
		page_size = 10
		glog.Error("ListPage page_size:", page_size, " <= 0")
	}
	return v.Limit(page_size).Offset((page - 1) * page_size)
}

func (v *PsqlDB) Begin() *PsqlDB {
	if atomic.CompareAndSwapInt32(&v.tx_flag, 0, 1) {
		v.DB = v.DB.Begin()
		return v
	} else {
		glog.Debug("PsqlDB is trasactions, ignore this time")
	}
	return nil
}
func (v *PsqlDB) Commit() *PsqlDB {
	if atomic.CompareAndSwapInt32(&v.tx_flag, 1, 0) {
		v.DB = v.DB.Commit()
		// 执行提交的代码
		if len(v.fn_after_commit) > 0 {
			for _, f := range v.fn_after_commit {
				f()
			}
		}
	} else {
		glog.Debug("PsqlDB is commited! ignore this time")
	}
	return v
}

func (v *PsqlDB) Rollback() *PsqlDB {
	if atomic.CompareAndSwapInt32(&v.tx_flag, 1, 0) {
		v.DB = v.DB.Rollback()
	} else {
		glog.Debug("PsqlDB is rollbacked! ignore this time")
	}
	return v
}

// 提交后执行的代码
func (v *PsqlDB) AfterCommit(f ...func()) (err error) {
	if 1 != atomic.LoadInt32(&v.tx_flag) {
		err = errors.New("is not tx db!")
		return
	}
	v.fn_after_commit = append(v.fn_after_commit, f...)
	return
}

func InitPsqlDb(psqlUrl string, debug bool) (*gorm.DB, error) {
	if db, ok := m_psqlDb[psqlUrl]; ok {
		return db, nil
	}

	db, err := gorm.Open("postgres", psqlUrl)
	if err != nil {
		glog.Fatalf("open postgresql(%v) failed! err: %v", psqlUrl, err)
		panic("open postgresql fail!")
	}
	fmt.Println("open sql ok")
	glog.Infof("open psql(%s) ok!", psqlUrl)

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	//db.SingularTable(true) // 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	db.LogMode(debug)

	m_psqlDb[psqlUrl] = db

	if m_db == nil {
		m_db = db
	}
	fmt.Println("open database success!")
	return db, nil
}

func InitMysqlDb(mysqlUrl string, debug bool) (*gorm.DB, error) {
	if db, ok := m_psqlDb[mysqlUrl]; ok {
		return db, nil
	}

	db, err := gorm.Open("mysql", mysqlUrl)
	if err != nil {
		glog.Fatalf("open mysql(%v) failed! err: %v", mysqlUrl, err)
		panic("open mysql fail!")
	}
	fmt.Println("open sql ok")
	glog.Infof("open mysql(%s) ok!", mysqlUrl)

	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	//db.SingularTable(true) // 如果设置为true,`User`的默认表名为`user`,使用`TableName`设置的表名不受影响
	db.LogMode(debug)

	m_psqlDb[mysqlUrl] = db

	if m_db == nil {
		m_db = db
	}
	fmt.Println("open database success!")
	return db, nil
}

// 获取一个事务db
func GetTxDB(c *gin.Context) *PsqlDB {
	var db *PsqlDB
	if c == nil {
		db = GetDB().Begin()
	} else {
		// 是否已经存在
		conn, exist := c.Get("sqldao")
		if exist {
			db, exist = conn.(*PsqlDB)
		}
		if !exist {
			db = GetDB().Begin() // 创建新事务db
			c.Set("sqldao", db)  // 关联到contex中
		}
	}
	return db
}

func GetDB() *PsqlDB {
	return &PsqlDB{DB: m_db, tx_flag: 0} // 返回默认的db
}
