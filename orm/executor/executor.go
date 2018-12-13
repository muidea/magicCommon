package executor

import "muidea.com/magicCommon/orm/mysql"

// Executor 数据库访问对象
type Executor interface {
	Release()
	BeginTransaction()
	Commit()
	Rollback()
	Query(sql string)
	Next() bool
	Finish()
	GetField(value ...interface{})
	Insert(sql string) int64
	Delete(sql string) int64
	Update(sql string) int64
	Execute(sql string)
}

// NewExecutor NewExecutor
func NewExecutor(user, password, address, dbName string) (Executor, error) {
	return mysql.Fetch(user, password, address, dbName)
}
