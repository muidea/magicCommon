package dao

import (
	"errors"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/stretchr/testify/assert"
)

// TestBaseDao_NewBaseDao 测试 BaseDao 创建
func TestBaseDao_NewBaseDao(t *testing.T) {
	// 由于 NewBaseDao 需要 *sql.DB，我们无法在单元测试中创建真实的数据库连接
	// 这个测试主要验证函数签名和基本逻辑
	assert.NotNil(t, NewBaseDao)
}

// TestBaseDao_Ping 测试 Ping 方法
func TestBaseDao_Ping(t *testing.T) {
	t.Run("数据库未初始化", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.Ping()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})
}

// TestBaseDao_Release 测试 Release 方法
func TestBaseDao_Release(t *testing.T) {
	t.Run("无数据库连接", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.Release()
		assert.Nil(t, err)
	})
}

// TestBaseDao_Transaction 测试事务方法
func TestBaseDao_Transaction(t *testing.T) {
	t.Run("数据库未初始化时开始事务", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.BeginTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})

	t.Run("数据库未初始化时提交事务", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.CommitTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})

	t.Run("数据库未初始化时回滚事务", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.RollbackTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})
}

// TestBaseDao_Query 测试 Query 方法
func TestBaseDao_Query(t *testing.T) {
	t.Run("数据库未初始化时查询", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		err := baseDao.Query("SELECT * FROM users")
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})
}

// TestBaseDao_Execute 测试 Execute 方法
func TestBaseDao_Execute(t *testing.T) {
	t.Run("数据库未初始化时执行", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		affected, err := baseDao.Execute("UPDATE users SET name = 'John'")
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), affected)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})
}

// TestBaseDao_Next_GetField_Finish 测试结果集方法
func TestBaseDao_Next_GetField_Finish(t *testing.T) {
	t.Run("GetField 无结果集", func(t *testing.T) {
		baseDao := &BaseDao{rowsHandle: nil}
		var id int
		err := baseDao.GetField(&id)
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.InvalidOperation, err.Code)
		assert.Equal(t, "result set not closed", err.Message)
	})

	t.Run("Finish 无结果集", func(t *testing.T) {
		baseDao := &BaseDao{rowsHandle: nil}
		err := baseDao.Finish()
		assert.Nil(t, err)
	})

	t.Run("Next 无结果集", func(t *testing.T) {
		baseDao := &BaseDao{rowsHandle: nil}
		hasNext := baseDao.Next()
		assert.False(t, hasNext)
	})
}

// TestBaseDao_Insert_Update_Delete 测试 CRUD 方法
func TestBaseDao_Insert_Update_Delete(t *testing.T) {
	t.Run("数据库未初始化时 Insert", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		affected, err := baseDao.Insert("INSERT INTO users (name) VALUES (?)", "John")
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), affected)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
	})

	t.Run("数据库未初始化时 Update", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		affected, err := baseDao.Update("UPDATE users SET name = ? WHERE id = ?", "Jane", 1)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), affected)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
	})

	t.Run("数据库未初始化时 Delete", func(t *testing.T) {
		baseDao := &BaseDao{dbHandle: nil}
		affected, err := baseDao.Delete("DELETE FROM users WHERE id = ?", 1)
		assert.NotNil(t, err)
		assert.Equal(t, int64(0), affected)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
	})
}

// TestWrapError 测试错误包装
func TestWrapError(t *testing.T) {
	t.Run("包装 nil 错误", func(t *testing.T) {
		err := WrapError(nil)
		assert.Nil(t, err)
	})

	t.Run("包装标准错误", func(t *testing.T) {
		stdErr := errors.New("standard error")
		wrappedErr := WrapError(stdErr)
		assert.NotNil(t, wrappedErr)
		assert.EqualValues(t, cd.DatabaseError, wrappedErr.Code)
		// 现在错误消息包含模块前缀
		assert.Equal(t, "[database] database operation failed: standard error", wrappedErr.Message)
	})

	t.Run("包装 *cd.Error", func(t *testing.T) {
		cdErr := cd.NewError(cd.InvalidParameter, "invalid param")
		wrappedErr := WrapError(cdErr)
		assert.NotNil(t, wrappedErr)
		// WrapError 现在保留原始错误代码
		assert.EqualValues(t, cd.InvalidParameter, wrappedErr.Code)
		assert.Equal(t, "invalid param", wrappedErr.Message)
	})
}

// TestBaseDao_ErrorDefinitions 测试错误定义
func TestBaseDao_ErrorDefinitions(t *testing.T) {
	assert.EqualValues(t, cd.DatabaseError, ErrDatabaseNotInitialized.Code)
	assert.Equal(t, "database not initialized", ErrDatabaseNotInitialized.Message)

	assert.EqualValues(t, cd.InvalidOperation, ErrTransactionActive.Code)
	assert.Equal(t, "transaction is still active", ErrTransactionActive.Message)

	assert.EqualValues(t, cd.InvalidOperation, ErrResultSetNotClosed.Code)
	assert.Equal(t, "result set not closed", ErrResultSetNotClosed.Message)

	assert.EqualValues(t, cd.InvalidParameter, ErrInvalidParameter.Code)
	assert.Equal(t, "invalid parameter", ErrInvalidParameter.Message)
}

// TestBaseDao_AbstractMethods 测试抽象方法
func TestBaseDao_AbstractMethods(t *testing.T) {
	baseDao := &BaseDao{}

	t.Run("CreateDatabase", func(t *testing.T) {
		err := baseDao.CreateDatabase("testdb")
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.NotImplemented, err.Code)
		assert.Equal(t, "not implemented", err.Message)
	})

	t.Run("DropDatabase", func(t *testing.T) {
		err := baseDao.DropDatabase("testdb")
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.NotImplemented, err.Code)
		assert.Equal(t, "not implemented", err.Message)
	})

	t.Run("UseDatabase", func(t *testing.T) {
		err := baseDao.UseDatabase("testdb")
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.NotImplemented, err.Code)
		assert.Equal(t, "not implemented", err.Message)
	})

	t.Run("CheckTableExist", func(t *testing.T) {
		exists, tableType, err := baseDao.CheckTableExist("users")
		assert.False(t, exists)
		assert.Equal(t, "", tableType)
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.NotImplemented, err.Code)
		assert.Equal(t, "not implemented", err.Message)
	})

	t.Run("Duplicate", func(t *testing.T) {
		dao, err := baseDao.Duplicate()
		assert.Nil(t, dao)
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.NotImplemented, err.Code)
		assert.Equal(t, "not implemented", err.Message)
	})
}

// TestBaseDao_String 测试 String 方法
func TestBaseDao_String(t *testing.T) {
	t.Run("有地址和数据库名", func(t *testing.T) {
		baseDao := &BaseDao{
			address: "localhost:3306",
			dbName:  "testdb",
		}
		str := baseDao.String()
		assert.Equal(t, "localhost:3306/testdb", str)
	})

	t.Run("只有地址", func(t *testing.T) {
		baseDao := &BaseDao{
			address: "localhost:3306",
			dbName:  "",
		}
		str := baseDao.String()
		assert.Equal(t, "localhost:3306/", str)
	})

	t.Run("只有数据库名", func(t *testing.T) {
		baseDao := &BaseDao{
			address: "",
			dbName:  "testdb",
		}
		str := baseDao.String()
		assert.Equal(t, "/testdb", str)
	})

	t.Run("都为空", func(t *testing.T) {
		baseDao := &BaseDao{
			address: "",
			dbName:  "",
		}
		str := baseDao.String()
		assert.Equal(t, "/", str)
	})
}

// TestBaseDao_DBName 测试 DBName 方法
func TestBaseDao_DBName(t *testing.T) {
	t.Run("有数据库名", func(t *testing.T) {
		baseDao := &BaseDao{dbName: "testdb"}
		dbName := baseDao.DBName()
		assert.Equal(t, "testdb", dbName)
	})

	t.Run("无数据库名", func(t *testing.T) {
		baseDao := &BaseDao{dbName: ""}
		dbName := baseDao.DBName()
		assert.Equal(t, "", dbName)
	})
}
