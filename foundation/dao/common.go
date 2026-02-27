package dao

import (
	"database/sql"
	"fmt"
	"sync"
	"sync/atomic"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"
	"log/slog"
)

// Dao 数据库访问对象接口
type Dao interface {
	DBName() string
	String() string
	Release() *cd.Error
	Ping() *cd.Error
	BeginTransaction() *cd.Error
	CommitTransaction() *cd.Error
	RollbackTransaction() *cd.Error
	CreateDatabase(dbName string) *cd.Error
	DropDatabase(dbName string) *cd.Error
	UseDatabase(dbName string) *cd.Error
	Query(sql string, args ...any) *cd.Error
	Next() bool
	Finish() *cd.Error
	GetField(value ...interface{}) *cd.Error
	Insert(sql string, args ...any) (int64, *cd.Error)
	Update(sql string, args ...any) (int64, *cd.Error)
	Delete(sql string, args ...any) (int64, *cd.Error)
	Execute(sql string, args ...any) (int64, *cd.Error)
	CheckTableExist(tableName string) (bool, string, *cd.Error)
	Duplicate() (Dao, *cd.Error)
}

// Driver 数据库驱动接口
type Driver interface {
	// Open 打开数据库连接
	Open(connectionString string) (*sql.DB, error)
	// Name 返回驱动名称
	Name() string
	// DefaultConnectionString 生成默认连接字符串
	DefaultConnectionString(user, password, address, dbName string) string
}

// 驱动注册表
var (
	drivers     = make(map[string]Driver)
	driversLock sync.RWMutex
)

// RegisterDriver 注册数据库驱动
func RegisterDriver(name string, driver Driver) {
	driversLock.Lock()
	defer driversLock.Unlock()

	if driver == nil {
		panic("dao: RegisterDriver driver is nil")
	}

	if _, dup := drivers[name]; dup {
		panic("dao: RegisterDriver called twice for driver " + name)
	}

	drivers[name] = driver
}

// GetDriver 获取已注册的驱动
func GetDriver(name string) (Driver, bool) {
	driversLock.RLock()
	defer driversLock.RUnlock()

	driver, ok := drivers[name]
	return driver, ok
}

// AvailableDrivers 返回所有可用的驱动名称
func AvailableDrivers() []string {
	driversLock.RLock()
	defer driversLock.RUnlock()

	names := make([]string, 0, len(drivers))
	for name := range drivers {
		names = append(names, name)
	}

	return names
}

// 错误定义
var (
	ErrDatabaseNotInitialized = cd.NewError(cd.DatabaseError, "database not initialized")
	ErrTransactionActive      = cd.NewError(cd.InvalidOperation, "transaction is still active")
	ErrResultSetNotClosed     = cd.NewError(cd.InvalidOperation, "result set not closed")
	ErrInvalidParameter       = cd.NewError(cd.InvalidParameter, "invalid parameter")
	ErrDriverNotFound         = cd.NewError(cd.DatabaseError, "database driver not found")
)

// BaseDao 基础DAO结构，包含所有公共字段
type BaseDao struct {
	dbHandle   *sql.DB
	dbTxCount  int32
	dbTx       *sql.Tx
	rowsHandle *sql.Rows
	user       string
	password   string
	address    string
	dbName     string
}

// DaoOption 配置选项函数类型
type DaoOption func(*BaseDao)

// WithUser 设置用户名
func WithUser(user string) DaoOption {
	return func(d *BaseDao) {
		d.user = user
	}
}

// WithPassword 设置密码
func WithPassword(password string) DaoOption {
	return func(d *BaseDao) {
		d.password = password
	}
}

// WithAddress 设置地址
func WithAddress(address string) DaoOption {
	return func(d *BaseDao) {
		d.address = address
	}
}

// WithDBName 设置数据库名称
func WithDBName(dbName string) DaoOption {
	return func(d *BaseDao) {
		d.dbName = dbName
	}
}

// NewBaseDao 创建基础DAO实例（新版本，支持 Functional Options）
func NewBaseDao(dbHandle *sql.DB, opts ...DaoOption) *BaseDao {
	dao := &BaseDao{
		dbHandle:   dbHandle,
		dbTxCount:  0,
		dbTx:       nil,
		rowsHandle: nil,
		// 默认值
		user:     "",
		password: "",
		address:  "",
		dbName:   "",
	}

	for _, opt := range opts {
		opt(dao)
	}

	return dao
}

// NewBaseDaoLegacy 旧版构造函数，保持向后兼容性
func NewBaseDaoLegacy(dbHandle *sql.DB, user, password, address, dbName string) *BaseDao {
	return NewBaseDao(dbHandle,
		WithUser(user),
		WithPassword(password),
		WithAddress(address),
		WithDBName(dbName),
	)
}

// WrapError 将标准错误包装为 *cd.Error
// 注意：为了性能考虑，默认不添加堆栈跟踪
// 如果需要堆栈跟踪，请使用 WrapErrorWithTrace
func WrapError(err error) *cd.Error {
	return util.DatabaseErrorFactory.Wrap(cd.DatabaseError, err, "database operation failed")
}

// logDatabaseError 记录数据库错误并返回包装后的错误
func logDatabaseError(operation, connectStr string, err error) *cd.Error {
	slog.Error("database operation failed",
		"operation", operation,
		"connectStr", connectStr,
		"error", err.Error())
	return WrapError(err)
}

// logSQLError 记录 SQL 执行错误并返回包装后的错误
func logSQLError(operation, sqlStr string, err error) *cd.Error {
	slog.Error("SQL operation failed",
		"operation", operation,
		"sql", sqlStr,
		"error", err)
	return WrapError(err)
}

// DBName 返回数据库名称
func (s *BaseDao) DBName() string {
	return s.dbName
}

// String 返回数据库连接字符串表示
func (s *BaseDao) String() string {
	return fmt.Sprintf("%s/%s", s.address, s.dbName)
}

// Release 释放数据库连接
func (s *BaseDao) Release() *cd.Error {
	if s.rowsHandle != nil {
		if err := s.rowsHandle.Close(); err != nil {
			slog.Error("failed to close rows handle", "error", err)
		}
	}
	s.rowsHandle = nil

	if s.dbHandle != nil {
		if err := s.dbHandle.Close(); err != nil {
			return WrapError(err)
		}
	}
	s.dbHandle = nil

	return nil
}

// Ping 检查数据库连接
func (s *BaseDao) Ping() *cd.Error {
	if s.dbHandle == nil {
		return ErrDatabaseNotInitialized
	}

	return WrapError(s.dbHandle.Ping())
}

// BeginTransaction 开始事务
func (s *BaseDao) BeginTransaction() *cd.Error {
	if s.dbHandle == nil {
		return ErrDatabaseNotInitialized
	}

	atomic.AddInt32(&s.dbTxCount, 1)
	if s.dbTx == nil && s.dbTxCount == 1 {
		if s.rowsHandle != nil {
			_ = s.rowsHandle.Close()
		}
		s.rowsHandle = nil

		tx, err := s.dbHandle.Begin()
		if err != nil {
			return WrapError(err)
		}

		s.dbTx = tx
	}

	return nil
}

// CommitTransaction 提交事务
func (s *BaseDao) CommitTransaction() *cd.Error {
	if s.dbHandle == nil {
		return ErrDatabaseNotInitialized
	}

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err := s.dbTx.Commit()
		s.dbTx = nil
		return WrapError(err)
	}

	return nil
}

// RollbackTransaction 回滚事务
func (s *BaseDao) RollbackTransaction() *cd.Error {
	if s.dbHandle == nil {
		return ErrDatabaseNotInitialized
	}

	atomic.AddInt32(&s.dbTxCount, -1)
	if s.dbTx != nil && s.dbTxCount == 0 {
		err := s.dbTx.Rollback()
		s.dbTx = nil
		return WrapError(err)
	}

	return nil
}

// Query 执行查询
func (s *BaseDao) Query(sqlStr string, args ...any) *cd.Error {
	if s.dbHandle == nil {
		return ErrDatabaseNotInitialized
	}

	if s.rowsHandle != nil {
		_ = s.rowsHandle.Close()
	}
	s.rowsHandle = nil

	var err error
	if s.dbTx != nil {
		s.rowsHandle, err = s.dbTx.Query(sqlStr, args...)
	} else {
		s.rowsHandle, err = s.dbHandle.Query(sqlStr, args...)
	}

	if err != nil {
		return logSQLError("execute query", sqlStr, err)
	}

	return nil
}

// Next 移动到下一行
func (s *BaseDao) Next() bool {
	if s.rowsHandle == nil {
		return false
	}

	return s.rowsHandle.Next()
}

// Finish 完成查询
func (s *BaseDao) Finish() *cd.Error {
	if s.rowsHandle == nil {
		return nil
	}

	err := s.rowsHandle.Close()
	s.rowsHandle = nil
	return WrapError(err)
}

// GetField 获取字段值
func (s *BaseDao) GetField(value ...interface{}) *cd.Error {
	if s.rowsHandle == nil {
		return ErrResultSetNotClosed
	}

	return WrapError(s.rowsHandle.Scan(value...))
}

// Insert 执行插入
func (s *BaseDao) Insert(sqlStr string, args ...any) (int64, *cd.Error) {
	return s.Execute(sqlStr, args...)
}

// Update 执行更新
func (s *BaseDao) Update(sqlStr string, args ...any) (int64, *cd.Error) {
	return s.Execute(sqlStr, args...)
}

// Delete 执行删除
func (s *BaseDao) Delete(sqlStr string, args ...any) (int64, *cd.Error) {
	return s.Execute(sqlStr, args...)
}

// Execute 执行SQL语句
func (s *BaseDao) Execute(sqlStr string, args ...any) (int64, *cd.Error) {
	if s.dbHandle == nil {
		return 0, ErrDatabaseNotInitialized
	}

	var result sql.Result
	var err error
	if s.dbTx != nil {
		result, err = s.dbTx.Exec(sqlStr, args...)
	} else {
		result, err = s.dbHandle.Exec(sqlStr, args...)
	}

	if err != nil {
		return 0, logSQLError("execute sql", sqlStr, err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return 0, logSQLError("get affected rows for sql", sqlStr, err)
	}

	return affectedRows, nil
}

// Duplicate 复制DAO实例
func (s *BaseDao) Duplicate() (Dao, *cd.Error) {
	// 这是一个抽象方法，需要在具体实现中重写
	return nil, cd.NewError(cd.NotImplemented, "not implemented")
}

// CreateDatabase 创建数据库（抽象方法，需要具体实现）
func (s *BaseDao) CreateDatabase(dbName string) *cd.Error {
	return cd.NewError(cd.NotImplemented, "not implemented")
}

// DropDatabase 删除数据库（抽象方法，需要具体实现）
func (s *BaseDao) DropDatabase(dbName string) *cd.Error {
	return cd.NewError(cd.NotImplemented, "not implemented")
}

// UseDatabase 使用数据库（抽象方法，需要具体实现）
func (s *BaseDao) UseDatabase(dbName string) *cd.Error {
	return cd.NewError(cd.NotImplemented, "not implemented")
}

// CheckTableExist 检查表是否存在（抽象方法，需要具体实现）
func (s *BaseDao) CheckTableExist(tableName string) (bool, string, *cd.Error) {
	return false, "", cd.NewError(cd.NotImplemented, "not implemented")
}
