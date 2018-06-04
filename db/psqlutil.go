package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jie123108/glog"
	_ "github.com/lib/pq"
	"gopkg.in/gorp.v1"
)

type PsqlDao struct {
	psqlUrl  string
	sqlDebug bool
	dbMap    *gorp.DbMap
	tx       *gorp.Transaction
	useTx    bool
}

var m_psqlDb map[string]*sql.DB = make(map[string]*sql.DB)

//var psqlDb *sql.DB

func InitPsqlDb(psqlUrl string) (*sql.DB, error) {
	if db, ok := m_psqlDb[psqlUrl]; ok {
		return db, nil
	}
	//	if psqlDb != nil {
	//		return psqlDb, nil
	//	}

	var err error
	psqlDb, err := sql.Open("postgres", psqlUrl)
	if err != nil {
		glog.Fatalf("open postgresql(%v) failed! err: %v", psqlUrl, err)
	}
	fmt.Println("open sql ok")
	m_psqlDb[psqlUrl] = psqlDb
	glog.Infof("open psql(%s) ok!", psqlUrl)
	return psqlDb, nil
}

func NewPsqlDao(psqlUrl string, sqlDebug bool, tableName string, tablesu interface{}, isautoid bool, keys []string) (dao *PsqlDao, err error) {
	dao = &PsqlDao{
		psqlUrl:  psqlUrl,
		sqlDebug: sqlDebug,
	}
	db, err := InitPsqlDb(psqlUrl)
	if err != nil {
		return
	}

	if dao.dbMap == nil {
		dao.dbMap = &gorp.DbMap{
			Db:      db,
			Dialect: gorp.PostgresDialect{},
		}
		if sqlDebug {
			var traceon = "["
			traceon += tableName
			traceon += "dao]"
			dao.dbMap.TraceOn(traceon, log.New(os.Stdout, "psql:", log.Lmicroseconds))
		}
	}

	if len(tableName) > 0 {
		dao.dbMap.AddTableWithName(tablesu, tableName).SetKeys(isautoid, keys...)
	}
	return
}

func (this *PsqlDao) AddTableWithName(tablestu interface{}, tablename string, isautoid bool, keys []string) {
	//glog.Errorf("addtablewithname(%v)", tablename)
	this.dbMap.AddTableWithName(tablestu, tablename).SetKeys(isautoid, keys...)
	return
}

func (this *PsqlDao) Begin() (err error) {
	this.tx, err = this.dbMap.Begin()
	this.useTx = true
	return
}

func (this *PsqlDao) Commit() (err error) {
	if this.useTx && this.tx != nil {
		err = this.tx.Commit()
	}
	return
}

func (this *PsqlDao) Rollback() (err error) {
	if this.useTx && this.tx != nil {
		err = this.tx.Rollback()
	}
	return
}

func (this *PsqlDao) Insert(obj interface{}) (err error) {
	if this.useTx {
		if this.tx == nil {
			return fmt.Errorf("this.tx is nil")
		}
		err = this.tx.Insert(obj)
	} else {
		if this.dbMap == nil {
			return fmt.Errorf("this.dbMap is nil")
		}
		err = this.dbMap.Insert(obj)
	}
	return
}

func (this *PsqlDao) Upsert(obj interface{}) (err error, effects int64, operator string) {
	if this.useTx {
		operator = "save"
		if this.tx == nil {
			return fmt.Errorf("this.dbMap is nil"), effects, operator
		}
		err = this.tx.Insert(obj)
		if err != nil && strings.Index(err.Error(), "#1062") > 0 {
			operator = "update"
			effects, err = this.tx.Update(obj)
		}
	} else {
		operator = "save"
		if this.dbMap == nil {
			return fmt.Errorf("this.dbMap is nil"), effects, operator
		}
		err = this.dbMap.Insert(obj)
		if err != nil && strings.Index(err.Error(), "#1062") > 0 {
			operator = "update"
			effects, err = this.dbMap.Update(obj)
		}
	}
	return err, effects, operator
}

func (this *PsqlDao) Update(obj interface{}) (err error, rows int64) {
	if this.useTx {
		if this.tx == nil {
			return fmt.Errorf("this.db is nil"), 0
		}
		rows, err = this.tx.Update(obj)
	} else {
		if this.dbMap == nil {
			return fmt.Errorf("this.db is nil"), 0
		}
		rows, err = this.dbMap.Update(obj)
	}
	return err, rows
}

func (this *PsqlDao) Exec(sqlstr string, args ...interface{}) (sql.Result, error) {
	if this.useTx {
		return this.tx.Exec(sqlstr, args...)
	}
	return this.dbMap.Exec(sqlstr, args...)
}

func (this *PsqlDao) Select(results interface{}, sql_select string, page int, page_size int, args ...interface{}) ([]interface{}, error) {
	offset := 0
	if page > 0 {
		offset = (page - 1) * page_size
	}
	if page_size > 0 {
		sql_select += fmt.Sprintf(" limit %v offset %v", page_size, offset)
	}

	if this.useTx {
		return this.tx.Select(results, sql_select, args...)
	}

	return this.dbMap.Select(results, sql_select, args...)

}

func (this *PsqlDao) SelectOne(result interface{}, sql_select string, args ...interface{}) error {
	if this.useTx {
		return this.tx.SelectOne(result, sql_select, args...)
	}
	return this.dbMap.SelectOne(result, sql_select, args...)
}

type PsqlCountStruct struct {
	Count int `json:"count" db:"count"`
}

func (this *PsqlDao) Count(sqlCount string) (int, error) {
	total := 0
	var err error
	results := []PsqlCountStruct{}
	_, err = this.Select(&results, sqlCount, 0, 0)

	if err == nil {
		if len(results) > 0 {
			total = results[0].Count
		}
	}

	return total, err
}
