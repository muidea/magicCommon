//go:build mysql
// +build mysql

package dao

import (
	"database/sql"
	"fmt"
	"time"

	cd "github.com/muidea/magicCommon/def"

	_ "github.com/go-sql-driver/mysql" //引入Mysql驱动
)

// mysqlDriver MySQL驱动实现
type mysqlDriver struct{}

func (d *mysqlDriver) Open(connectionString string) (*sql.DB, error) {
	return sql.Open("mysql", connectionString)
}

func (d *mysqlDriver) Name() string {
	return "mysql"
}

func (d *mysqlDriver) DefaultConnectionString(user, password, address, dbName string) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, address, dbName)
}

// init 注册MySQL驱动
func init() {
	RegisterDriver("mysql", &mysqlDriver{})
}

const (
	BaseTable = "BASE TABLE"
	View      = "VIEW"
)

type impl struct {
	*BaseDao
}

// Fetch 获取一个数据访问对象（使用默认MySQL驱动）
func Fetch(user, password, address, dbName string) (Dao, *cd.Error) {
	return FetchWithDriver("mysql", user, password, address, dbName)
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
	// 生产环境推荐配置：
	// - SetMaxOpenConns: 根据应用负载调整，通常为 (核心数 * 2) + 有效磁盘数
	// - SetMaxIdleConns: 设置为与SetMaxOpenConns相同或略小，避免频繁创建连接
	// - SetConnMaxLifetime: 设置合理生命周期，避免长时间占用连接
	// - SetConnMaxIdleTime: 设置空闲超时，及时释放不用的连接
	db.SetMaxOpenConns(25)                  // 最大打开连接数
	db.SetMaxIdleConns(10)                  // 最大空闲连接数
	db.SetConnMaxLifetime(30 * time.Minute) // 连接最大生命周期
	db.SetConnMaxIdleTime(5 * time.Minute)  // 连接最大空闲时间

	err = db.Ping()
	if err != nil {
		return nil, logDatabaseError("ping database", connectStr, err)
	}

	baseDao := NewBaseDaoLegacy(db, user, password, address, dbName)
	return &impl{BaseDao: baseDao}, nil
}

// CreateDatabase 创建数据库
func (s *impl) CreateDatabase(dbName string) *cd.Error {
	_, err := s.Execute(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

// DropDatabase 删除数据库
func (s *impl) DropDatabase(dbName string) *cd.Error {
	_, err := s.Execute(fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

// UseDatabase 使用数据库
func (s *impl) UseDatabase(dbName string) *cd.Error {
	s.dbName = dbName
	_, err := s.Execute(fmt.Sprintf("USE `%s`", dbName))
	if err != nil {
		return err
	}
	_, err = s.Execute("FLUSH TABLES")
	return err
}

// CheckTableExist 检查表是否存在
func (s *impl) CheckTableExist(tableName string) (bool, string, *cd.Error) {
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

// Duplicate 复制DAO实例
func (s *impl) Duplicate() (Dao, *cd.Error) {
	return Fetch(s.user, s.password, s.address, s.dbName)
}
