//go:build !mysql
// +build !mysql

package dao

import (
	"errors"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/stretchr/testify/assert"
)

// TestErrorHandlingCompleteness 错误处理完整性测试
func TestErrorHandlingCompleteness(t *testing.T) {
	t.Run("错误分类验证", func(t *testing.T) {
		// 验证所有预定义错误都有正确的分类
		testCases := []struct {
			name     string
			err      *cd.Error
			expected cd.Code
		}{
			{"数据库未初始化", ErrDatabaseNotInitialized, cd.DatabaseError},
			{"事务活跃", ErrTransactionActive, cd.InvalidOperation},
			{"结果集未关闭", ErrResultSetNotClosed, cd.InvalidOperation},
			{"无效参数", ErrInvalidParameter, cd.InvalidParameter},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				assert.EqualValues(t, tc.expected, tc.err.Code)
				assert.NotEmpty(t, tc.err.Message)
			})
		}
	})

	t.Run("错误消息一致性", func(t *testing.T) {
		// 验证相同错误条件的错误消息一致
		baseDao1 := &BaseDao{dbHandle: nil}
		baseDao2 := &BaseDao{dbHandle: nil}

		err1 := baseDao1.Ping()
		err2 := baseDao2.Ping()

		assert.NotNil(t, err1)
		assert.NotNil(t, err2)
		assert.Equal(t, err1.Code, err2.Code)
		assert.Equal(t, err1.Message, err2.Message)
	})

	t.Run("错误链支持验证", func(t *testing.T) {
		// 测试错误包装和链式错误
		originalErr := errors.New("original database error")

		// 第一层包装
		wrappedErr1 := WrapError(originalErr)
		assert.NotNil(t, wrappedErr1)
		assert.EqualValues(t, cd.DatabaseError, wrappedErr1.Code)
		assert.Contains(t, wrappedErr1.Message, "original database error")

		// 第二层包装（应该保留原始错误代码）
		wrappedErr2 := WrapError(wrappedErr1)
		assert.NotNil(t, wrappedErr2)
		assert.EqualValues(t, cd.DatabaseError, wrappedErr2.Code)
		assert.Contains(t, wrappedErr2.Message, "original database error")
	})

	t.Run("nil错误处理", func(t *testing.T) {
		// 验证所有可能返回错误的方法都能正确处理nil错误
		assert.Nil(t, WrapError(nil))

		baseDao := &BaseDao{rowsHandle: nil}
		assert.Nil(t, baseDao.Finish())

		// Release 方法在dbHandle为nil时应返回nil
		baseDao.dbHandle = nil
		assert.Nil(t, baseDao.Release())
	})

	t.Run("错误上下文信息", func(t *testing.T) {
		// 验证错误包含足够的上下文信息
		baseDao := &BaseDao{
			address:  "localhost:5432",
			dbName:   "testdb",
			dbHandle: nil,
		}

		err := baseDao.Ping()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)

		// 错误应该能够通过String()方法获取更多上下文
		daoString := baseDao.String()
		assert.Contains(t, daoString, "localhost:5432")
		assert.Contains(t, daoString, "testdb")
	})
}

// TestErrorRecoveryMechanisms 错误恢复机制测试
func TestErrorRecoveryMechanisms(t *testing.T) {
	t.Run("连接失败恢复", func(t *testing.T) {
		// 测试数据库连接失败时的错误处理
		// 由于需要模拟数据库连接失败，这里只验证错误类型
		baseDao := &BaseDao{dbHandle: nil}

		err := baseDao.Ping()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)

		// 验证错误消息清晰
		assert.Equal(t, "database not initialized", err.Message)
	})

	t.Run("事务失败回滚", func(t *testing.T) {
		// 测试事务相关错误处理
		baseDao := &BaseDao{dbHandle: nil}

		// 开始事务失败
		err := baseDao.BeginTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)

		// 提交事务失败
		err = baseDao.CommitTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)

		// 回滚事务失败
		err = baseDao.RollbackTransaction()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
	})

	t.Run("查询失败处理", func(t *testing.T) {
		// 测试查询相关错误处理
		baseDao := &BaseDao{dbHandle: nil}

		// 查询失败
		err := baseDao.Query("SELECT * FROM non_existent_table")
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)

		// 执行失败
		affected, err := baseDao.Execute("INVALID SQL STATEMENT")
		assert.Equal(t, int64(0), affected)
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
	})

	t.Run("结果集错误处理", func(t *testing.T) {
		baseDao := &BaseDao{rowsHandle: nil}

		// GetField 在无结果集时应返回错误
		var value int
		err := baseDao.GetField(&value)
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.InvalidOperation, err.Code)
		assert.Equal(t, "result set not closed", err.Message)

		// Next 在无结果集时应返回false但不报错
		hasNext := baseDao.Next()
		assert.False(t, hasNext)

		// Finish 在无结果集时应成功
		err = baseDao.Finish()
		assert.Nil(t, err)
	})
}

// TestErrorLoggingValidation 错误日志验证
func TestErrorLoggingValidation(t *testing.T) {
	t.Run("错误日志结构化", func(t *testing.T) {
		// 验证错误对象能够提供结构化的错误信息
		testCases := []struct {
			name     string
			err      *cd.Error
			expected string
		}{
			{
				name:     "数据库错误",
				err:      ErrDatabaseNotInitialized,
				expected: "code:8, message:database not initialized",
			},
			{
				name:     "无效操作错误",
				err:      ErrTransactionActive,
				expected: "code:21, message:transaction is still active",
			},
			{
				name:     "无效参数错误",
				err:      ErrInvalidParameter,
				expected: "code:3, message:invalid parameter",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				errorString := tc.err.Error()
				assert.Contains(t, errorString, tc.expected)
			})
		}
	})

	t.Run("错误频率统计", func(t *testing.T) {
		// 验证相同错误多次出现时，错误对象的一致性
		baseDao := &BaseDao{dbHandle: nil}

		errors := make([]*cd.Error, 10)
		for i := 0; i < 10; i++ {
			errors[i] = baseDao.Ping()
		}

		// 所有错误应该相同
		for i := 1; i < len(errors); i++ {
			assert.Equal(t, errors[0].Code, errors[i].Code)
			assert.Equal(t, errors[0].Message, errors[i].Message)
		}
	})
}

// TestWrapErrorFunctionValidation WrapError函数验证
func TestWrapErrorFunctionValidation(t *testing.T) {
	t.Run("标准错误包装", func(t *testing.T) {
		testCases := []struct {
			name         string
			inputErr     error
			expectedCode cd.Code
		}{
			{
				name:         "连接错误",
				inputErr:     errors.New("connection refused"),
				expectedCode: cd.DatabaseError,
			},
			{
				name:         "超时错误",
				inputErr:     errors.New("timeout"),
				expectedCode: cd.DatabaseError,
			},
			{
				name:         "语法错误",
				inputErr:     errors.New("syntax error"),
				expectedCode: cd.DatabaseError,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				wrappedErr := WrapError(tc.inputErr)
				assert.NotNil(t, wrappedErr)
				assert.EqualValues(t, tc.expectedCode, wrappedErr.Code)
				assert.Contains(t, wrappedErr.Message, tc.inputErr.Error())
			})
		}
	})

	t.Run("cd.Error包装", func(t *testing.T) {
		testCases := []struct {
			name         string
			inputErr     *cd.Error
			expectedCode cd.Code
		}{
			{
				name:         "数据库错误",
				inputErr:     cd.NewError(cd.DatabaseError, "database error"),
				expectedCode: cd.DatabaseError,
			},
			{
				name:         "无效参数错误",
				inputErr:     cd.NewError(cd.InvalidParameter, "invalid param"),
				expectedCode: cd.InvalidParameter,
			},
			{
				name:         "未授权错误",
				inputErr:     cd.NewError(cd.Unauthorized, "unauthorized"),
				expectedCode: cd.Unauthorized,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				wrappedErr := WrapError(tc.inputErr)
				assert.NotNil(t, wrappedErr)
				assert.EqualValues(t, tc.expectedCode, wrappedErr.Code)
				assert.Equal(t, tc.inputErr.Message, wrappedErr.Message)
			})
		}
	})

	t.Run("性能验证", func(t *testing.T) {
		// 验证WrapError函数性能
		err := errors.New("test error")

		// 多次调用不应该有性能问题
		for i := 0; i < 1000; i++ {
			wrappedErr := WrapError(err)
			assert.NotNil(t, wrappedErr)
		}
	})
}

// TestErrorHandlingIntegration 错误处理集成测试
func TestErrorHandlingIntegration(t *testing.T) {
	t.Run("完整错误处理流程", func(t *testing.T) {
		// 模拟一个完整的数据库操作流程中的错误处理
		baseDao := &BaseDao{dbHandle: nil}

		// 1. 连接失败
		err := baseDao.Ping()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)

		// 2. 尝试开始事务（应该失败）
		err = baseDao.BeginTransaction()
		assert.NotNil(t, err)

		// 3. 尝试查询（应该失败）
		err = baseDao.Query("SELECT 1")
		assert.NotNil(t, err)

		// 4. 尝试执行（应该失败）
		affected, err := baseDao.Execute("INSERT INTO test VALUES (1)")
		assert.Equal(t, int64(0), affected)
		assert.NotNil(t, err)

		// 5. 释放资源（应该成功）
		err = baseDao.Release()
		assert.Nil(t, err)
	})

	t.Run("错误传播验证", func(t *testing.T) {
		// 验证错误能够在调用链中正确传播
		func2 := func() *cd.Error {
			baseDao := &BaseDao{dbHandle: nil}
			return baseDao.Ping()
		}

		func1 := func() *cd.Error {
			return func2()
		}

		err := func1()
		assert.NotNil(t, err)
		assert.EqualValues(t, cd.DatabaseError, err.Code)
		assert.Equal(t, "database not initialized", err.Message)
	})
}
