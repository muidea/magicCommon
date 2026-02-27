//go:build !mysql
// +build !mysql

package dao

import (
	"database/sql"
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"

	_ "github.com/lib/pq" //引入PostgreSQL驱动
)

// postgresDriver PostgreSQL驱动实现
type postgresDriver struct{}

func (d *postgresDriver) Open(connectionString string) (*sql.DB, error) {
	return sql.Open("postgres", connectionString)
}

func (d *postgresDriver) Name() string {
	return "postgres"
}

func (d *postgresDriver) DefaultConnectionString(user, password, address, dbName string) string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", user, password, address, dbName)
}

// init 注册PostgreSQL驱动
func init() {
	RegisterDriver("postgres", &postgresDriver{})
}

const (
	BaseTable = "BASE TABLE"
	View      = "VIEW"
)

type impl struct {
	*BaseDao
}

// Fetch 获取一个数据访问对象（使用默认PostgreSQL驱动）
func Fetch(user, password, address, dbName string) (Dao, *cd.Error) {
	return FetchWithDriver("postgres", user, password, address, dbName)
}

// FetchWithDriver 使用指定驱动获取数据访问对象
func FetchWithDriver(driverName, user, password, address, dbName string) (Dao, *cd.Error) {
	driver, ok := GetDriver(driverName)
	if !ok {
		return nil, cd.NewError(cd.DatabaseError, fmt.Sprintf("database driver '%s' not found", driverName))
	}

	connectStr := driver.DefaultConnectionString(user, password, address, dbName)
	db, err := driver.Open(connectStr)
	if err != nil {
		return nil, logDatabaseError("open database", connectStr, err)
	}

	// 配置连接池优化参数
	// PostgreSQL连接池配置建议：
	// - SetMaxOpenConns: PostgreSQL对并发连接有限制，需要根据max_connections调整
	// - SetMaxIdleConns: 保持一定数量的空闲连接以提高性能
	// - SetConnMaxLifetime: PostgreSQL连接相对稳定，可以设置较长的生命周期
	// - SetConnMaxIdleTime: 及时释放长时间空闲的连接
	db.SetMaxOpenConns(25)                  // 最大打开连接数
	db.SetMaxIdleConns(10)                  // 最大空闲连接数
	db.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
	db.SetConnMaxIdleTime(10 * time.Minute) // 连接最大空闲时间

	err = db.Ping()
	if err != nil {
		return nil, logDatabaseError("ping database", connectStr, err)
	}

	baseDao := NewBaseDaoLegacy(db, user, password, address, dbName)
	return &impl{BaseDao: baseDao}, nil
}

// CreateDatabase 创建数据库
func (s *impl) CreateDatabase(dbName string) *cd.Error {
	_, err := s.Execute(fmt.Sprintf("CREATE DATABASE \"%s\"", dbName))
	return err
}

// DropDatabase 删除数据库
func (s *impl) DropDatabase(dbName string) *cd.Error {
	_, err := s.Execute(fmt.Sprintf("DROP DATABASE IF EXISTS \"%s\"", dbName))
	return err
}

// UseDatabase 使用数据库
func (s *impl) UseDatabase(dbName string) *cd.Error {
	s.dbName = dbName
	// PostgreSQL 不需要 USE 语句，连接时已经指定了数据库
	return nil
}

// CheckTableExist 检查表是否存在
func (s *impl) CheckTableExist(tableName string) (bool, string, *cd.Error) {
	sqlStr := fmt.Sprintf("SELECT tablename, CASE WHEN schemaname = 'public' THEN 'BASE TABLE' ELSE 'VIEW' END FROM pg_tables WHERE tablename ='%s' UNION ALL SELECT viewname, 'VIEW' FROM pg_views WHERE viewname ='%s'", tableName, tableName)

	err := s.Query(sqlStr)
	if err != nil {
		return false, "", err
	}

	defer func() { _ = s.Finish() }()

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

// Duplicate 复制DAO实例
func (s *impl) Duplicate() (Dao, *cd.Error) {
	return Fetch(s.user, s.password, s.address, s.dbName)
}
