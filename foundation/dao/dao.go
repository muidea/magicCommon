package dao

import (
	"database/sql"
	"fmt"
	"sync/atomic"

	"github.com/muidea/magicCommon/foundation/log"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动
)

const (
	BaseTable = "BASE TABLE"
	View      = "VIEW"
)

// Dao 数据库访问对象
type Dao interface {
	DBName() string
	String() string
	Release()
	Ping() error
	BeginTransaction() error
	CommitTransaction() error
	RollbackTransaction() error
	CreateDatabase(dbName string) error
	DropDatabase(dbName string) error
	UseDatabase(dbName string) error
	Query(sql string, args ...any) error
	Next() bool
	Finish()
	GetField(value ...interface{}) error
	Insert(sql string, args ...any) (int64, error)
	Update(sql string, args ...any) (int64, error)
	Delete(sql string, args ...any) (int64, error)
	Execute(sql string, args ...any) (int64, error)
	CheckTableExist(tableName string) (bool, string, error)
	Duplicate() (Dao, error)
}

type impl struct {
	dbHandle   *sql.DB
	dbTxCount  int32
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
	user       string
	password   string
	address    string
	dbName     string
	charSet    string
}

// Fetch 获取一个数据访问对象
func Fetch(user, password, address, dbName, charSet string) (Dao, error) {
	if charSet == "" {
		charSet = "utf8"
	}
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s", user, password, address, dbName, charSet)

	i := impl{dbHandle: nil, dbTx: nil, rowsHandle: nil, user: user, password: password, address: address, dbName: dbName, charSet: charSet}
	db, err := sql.Open("mysql", connectStr)
	if err != nil {
		log.Errorf("open database exception, err:%s", err.Error())
		return nil, err
	}

	i.dbHandle = db

	err = db.Ping()
	if err != nil {
		log.Errorf("ping database failed, err:%s", err.Error())
		return nil, err
	}

	return &i, err
}

func (s *impl) DBName() string {
	return s.dbName
}

func (s *impl) String() string {
	return fmt.Sprintf("%s/%s", s.address, s.dbName)
}

func (s *impl) Release() {
	if s.dbTx != nil {
		panic("dbTx isn't nil")
	}

	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbHandle != nil {
		_ = s.dbHandle.Close()
	}
	s.dbHandle = nil
}

func (s *impl) Ping() error {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	return s.dbHandle.Ping()
}

func (s *impl) BeginTransaction() error {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, err := s.dbHandle.Begin()
		if err != nil {
			return err
		}

		s.dbTx = tx
	}

	return nil
}

func (s *impl) CommitTransaction() error {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err := s.dbTx.Commit()
		if err != nil {
			s.dbTx = nil
			return err
		}

		s.dbTx = nil
	}

	return nil
}

func (s *impl) RollbackTransaction() error {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err := s.dbTx.Rollback()
		if err != nil {
			s.dbTx = nil
			return err
		}

		s.dbTx = nil
	}

	return nil
}

func (s *impl) CreateDatabase(dbName string) error {
	_, err := s.Execute(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

func (s *impl) DropDatabase(dbName string) error {
	_, err := s.Execute(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

func (s *impl) UseDatabase(dbName string) error {
	s.dbName = dbName
	_, err := s.Execute(fmt.Sprintf("USE `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

func (s *impl) Query(sql string, args ...any) error {
	if s.dbHandle == nil {
		panic("dbHanlde is nil")
	}

	if s.dbTx == nil {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, err := s.dbHandle.Query(sql, args...)
		if err != nil {
			return err
		}
		s.rowsHandle = rows
	} else {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, err := s.dbTx.Query(sql, args...)
		if err != nil {
			return err
		}
		s.rowsHandle = rows
	}

	return nil
}

func (s *impl) Next() bool {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	ret := s.rowsHandle.Next()
	if !ret {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}

	return ret
}

func (s *impl) Finish() {
	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

func (s *impl) GetField(value ...interface{}) error {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	err := s.rowsHandle.Scan(value...)
	return err
}

func (s *impl) Insert(sql string, args ...any) (int64, error) {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		execVal, execErr := s.dbHandle.Exec(sql, args...)
		if execErr != nil {
			return 0, execErr
		}

		idVal, idErr := execVal.LastInsertId()
		if idErr != nil {
			return 0, idErr
		}

		return idVal, nil
	}

	execVal, execErr := s.dbTx.Exec(sql, args...)
	if execErr != nil {
		return 0, execErr
	}

	idVal, idErr := execVal.LastInsertId()
	if idErr != nil {
		return 0, idErr
	}

	return idVal, nil
}

func (s *impl) Update(sql string, args ...any) (int64, error) {
	return s.Execute(sql, args...)
}

func (s *impl) Delete(sql string, args ...any) (int64, error) {
	return s.Execute(sql, args...)
}

func (s *impl) Execute(sql string, args ...any) (int64, error) {
	if s.dbHandle == nil {
		panic("dbHandle is nil")
	}

	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		execVal, execErr := s.dbHandle.Exec(sql, args...)
		if execErr != nil {
			return 0, execErr
		}

		rowNum, rowErr := execVal.RowsAffected()
		if rowErr != nil {
			return 0, rowErr
		}

		return rowNum, nil
	}

	execVal, execErr := s.dbTx.Exec(sql, args...)
	if execErr != nil {
		return 0, execErr
	}

	rowNum, rowErr := execVal.RowsAffected()
	if rowErr != nil {
		return 0, rowErr
	}

	return rowNum, nil
}

func (s *impl) CheckTableExist(tableName string) (bool, string, error) {
	sqlStr := fmt.Sprintf("SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_NAME ='%s' and TABLE_SCHEMA ='%s'", tableName, s.dbName)

	err := s.Query(sqlStr)
	if err != nil {
		return false, "", err
	}

	defer s.Finish()

	var tableNameVal, tableTypeVal sql.NullString
	if s.Next() {
		err = s.GetField(&tableNameVal, &tableTypeVal)
		if err != nil {
			return false, "", err
		}

		return true, tableTypeVal.String, nil
	}

	return false, "", nil
}

func (s *impl) Duplicate() (Dao, error) {
	return Fetch(s.user, s.password, s.address, s.dbName, s.charSet)
}
