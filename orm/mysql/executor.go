package mysql

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动
)

// Executor Executor
type Executor struct {
	dbHandle   *sql.DB
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
	dbName     string
}

// Fetch 获取一个数据访问对象
func Fetch(user, password, address, dbName string) (*Executor, error) {
	connectStr := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8", user, password, address, dbName)

	i := Executor{dbHandle: nil, dbTx: nil, rowsHandle: nil, dbName: dbName}
	db, err := sql.Open("mysql", connectStr)
	if err != nil {
		log.Printf("open database exception, err:%s", err.Error())
		return nil, err
	}

	//log.Print("open database connection...")
	i.dbHandle = db

	err = db.Ping()
	if err != nil {
		log.Printf("ping database failed, err:%s", err.Error())
		return nil, err
	}

	return &i, err
}

// Release Release
func (s *Executor) Release() {
	if s.dbTx != nil {
		panic("dbTx isn't nil")
	}

	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbHandle != nil {
		//log.Print("close database connection...")

		s.dbHandle.Close()
	}
	s.dbHandle = nil

}

// BeginTransaction BeginTransaction
func (s *Executor) BeginTransaction() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	tx, err := s.dbHandle.Begin()
	if err != nil {
		panic("begin transaction exception, err:" + err.Error())
	}

	s.dbTx = tx
	//log.Print("BeginTransaction")
}

// Commit Commit
func (s *Executor) Commit() {
	if s.dbTx == nil {
		panic("dbTx is nil")
	}

	err := s.dbTx.Commit()
	if err != nil {
		s.dbTx = nil

		panic("commit transaction exception, err:" + err.Error())
	}

	s.dbTx = nil
	//log.Print("Commit")
}

// Rollback Rollback
func (s *Executor) Rollback() {
	if s.dbTx == nil {
		panic("dbTx is nil")
	}

	err := s.dbTx.Rollback()
	if err != nil {
		s.dbTx = nil

		panic("rollback transaction exception, err:" + err.Error())
	}

	s.dbTx = nil
	//log.Print("Rollback")
}

// Query Query
func (s *Executor) Query(sql string) {
	//log.Printf("Query, sql:%s", sql)
	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHanlde is nil")
		}
		if s.rowsHandle != nil {
			s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, err := s.dbHandle.Query(sql)
		if err != nil {
			panic("query exception, err:" + err.Error() + ", sql:" + sql)
		}
		s.rowsHandle = rows
	} else {

		if s.rowsHandle != nil {
			s.rowsHandle.Close()
			s.rowsHandle = nil
		}

		rows, err := s.dbTx.Query(sql)
		if err != nil {
			panic("query exception, err:" + err.Error() + ", sql:" + sql)
		}
		s.rowsHandle = rows
	}
}

// Next Next
func (s *Executor) Next() bool {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	ret := s.rowsHandle.Next()
	if !ret {
		//log.Print("Next, close rows")
		s.rowsHandle.Close()
		s.rowsHandle = nil
	}

	return ret
}

// Finish Finish
func (s *Executor) Finish() {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
		s.rowsHandle = nil
	}
}

// GetField GetField
func (s *Executor) GetField(value ...interface{}) {
	if s.rowsHandle == nil {
		panic("rowsHandle is nil")
	}

	err := s.rowsHandle.Scan(value...)
	if err != nil {
		panic("scan exception, err:" + err.Error())
	}
}

// Insert Insert
func (s *Executor) Insert(sql string) int64 {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, err := s.dbHandle.Exec(sql)
		if err != nil {
			panic("exec exception, err:" + err.Error() + ", sql:" + sql)
		}

		idNum, err := result.LastInsertId()
		if err != nil {
			panic("insert failed exception, err:" + err.Error())
		}

		return idNum
	}

	result, err := s.dbTx.Exec(sql)
	if err != nil {
		panic("exec exception, err:" + err.Error() + ", sql:" + sql)
	}

	idNum, err := result.LastInsertId()
	if err != nil {
		panic("insert failed exception, err:" + err.Error())
	}

	return idNum
}

// Update Update
func (s *Executor) Update(sql string) int64 {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, err := s.dbHandle.Exec(sql)
		if err != nil {
			panic("exec exception, err:" + err.Error() + ", sql:" + sql)
		}

		num, err := result.RowsAffected()
		if err != nil {
			panic("rows affected exception, err:" + err.Error())
		}

		return num
	}

	result, err := s.dbTx.Exec(sql)
	if err != nil {
		panic("exec exception, err:" + err.Error() + ", sql:" + sql)
	}

	num, err := result.RowsAffected()
	if err != nil {
		panic("rows affected exception, err:" + err.Error())
	}

	return num
}

// Delete Delete
func (s *Executor) Delete(sql string) int64 {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, err := s.dbHandle.Exec(sql)
		if err != nil {
			panic("exec exception, err:" + err.Error() + ", sql:" + sql)
		}

		num, err := result.RowsAffected()
		if err != nil {
			panic("rows affected exception, err:" + err.Error())
		}

		return num
	}

	result, err := s.dbTx.Exec(sql)
	if err != nil {
		panic("exec exception, err:" + err.Error() + ", sql:" + sql)
	}

	num, err := result.RowsAffected()
	if err != nil {
		panic("rows affected exception, err:" + err.Error())
	}

	return num
}

// Execute Execute
func (s *Executor) Execute(sql string) {
	if s.rowsHandle != nil {
		s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	if s.dbTx == nil {
		if s.dbHandle == nil {
			panic("dbHandle is nil")
		}

		result, err := s.dbHandle.Exec(sql)
		if err != nil {
			panic("exec exception, err:" + err.Error() + ", sql:" + sql)
		}

		_, err = result.RowsAffected()
		if err != nil {
			panic("rows affected exception, err:" + err.Error())
		}

		return
	}

	result, err := s.dbTx.Exec(sql)
	if err != nil {
		panic("exec exception, err:" + err.Error() + ", sql:" + sql)
	}

	_, err = result.RowsAffected()
	if err != nil {
		panic("rows affected exception, err:" + err.Error())
	}
}

// CheckTableExist CheckTableExist
func (s *Executor) CheckTableExist(tableName string) (ret bool) {
	sql := fmt.Sprintf("SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_NAME ='%s' and TABLE_SCHEMA ='%s'", tableName, s.dbName)

	s.Query(sql)
	if s.Next() {
		ret = true
	} else {
		ret = false
	}
	s.Finish()

	return
}
