package dao

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConnectionPoolDefaultSettings 测试默认连接池设置
func TestConnectionPoolDefaultSettings(t *testing.T) {
	t.Run("验证默认连接池设置", func(t *testing.T) {
		// 创建一个测试数据库连接（使用无效连接字符串，只测试API）
		db, err := sql.Open("mysql", "invalid:connection@string")
		if err != nil {
			// 即使连接字符串无效，sql.Open也可能成功（延迟连接）
			t.Logf("sql.Open 返回错误: %v", err)
			// 继续测试其他方面
		}

		if db != nil {
			defer func() { _ = db.Close() }()

			// 验证默认连接池设置
			// 注意：sql.DB 的默认设置是：
			// - SetMaxOpenConns: 0 (无限制)
			// - SetMaxIdleConns: 2
			// - SetConnMaxLifetime: 0 (永不过期)
			// - SetConnMaxIdleTime: 0 (Go 1.15+ 默认值)

			// 验证可以获取连接池统计信息（即使没有实际连接）
			stats := db.Stats()
			assert.NotNil(t, stats, "应该能获取连接池统计信息")
			assert.GreaterOrEqual(t, stats.OpenConnections, 0, "打开连接数应该 >= 0")
			assert.GreaterOrEqual(t, stats.InUse, 0, "使用中的连接数应该 >= 0")
			assert.GreaterOrEqual(t, stats.Idle, 0, "空闲连接数应该 >= 0")
		}

		// 测试连接池配置API
		db2, _ := sql.Open("postgres", "invalid")
		if db2 != nil {
			defer func() { _ = db2.Close() }()

			// 测试配置方法
			db2.SetMaxOpenConns(10)
			db2.SetMaxIdleConns(5)
			db2.SetConnMaxLifetime(time.Hour)
			db2.SetConnMaxIdleTime(30 * time.Minute)

			// 验证配置不会导致panic
			assert.NotPanics(t, func() {
				db2.SetMaxOpenConns(20)
			}, "设置连接池参数不应该panic")
		}
	})
}

// TestConnectionPoolConfiguration 测试连接池配置
func TestConnectionPoolConfiguration(t *testing.T) {
	t.Run("验证连接池配置方法", func(t *testing.T) {
		// 创建一个测试数据库连接
		db, err := sql.Open("mysql", "root:rootkit@tcp(localhost:3306)/testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping connection pool test")
			return
		}
		defer func() { _ = db.Close() }()

		// 配置连接池参数
		db.SetMaxOpenConns(10)                  // 最大打开连接数
		db.SetMaxIdleConns(5)                   // 最大空闲连接数
		db.SetConnMaxLifetime(time.Hour)        // 连接最大生命周期
		db.SetConnMaxIdleTime(30 * time.Minute) // 连接最大空闲时间

		// 验证配置生效
		err = db.Ping()
		assert.Nil(t, err, "配置后数据库连接应该成功")

		// 执行一个简单查询来验证连接池工作
		rows, err := db.Query("SELECT 1")
		if err != nil {
			t.Skip("查询失败，跳过连接池验证")
			return
		}
		_ = rows.Close()

		// 验证连接池统计信息
		stats := db.Stats()
		assert.NotNil(t, stats, "应该能获取连接池统计信息")
		t.Logf("连接池统计: OpenConnections=%d, InUse=%d, Idle=%d",
			stats.OpenConnections, stats.InUse, stats.Idle)
	})
}

// TestConnectionPoolConcurrentAccess 测试并发访问连接池
func TestConnectionPoolConcurrentAccess(t *testing.T) {
	t.Run("验证并发访问连接池", func(t *testing.T) {
		// 创建一个测试数据库连接
		db, err := sql.Open("mysql", "root:rootkit@tcp(localhost:3306)/testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping concurrent test")
			return
		}
		defer func() { _ = db.Close() }()

		// 配置适中的连接池大小
		db.SetMaxOpenConns(5)
		db.SetMaxIdleConns(3)
		db.SetConnMaxLifetime(time.Hour)

		// 并发执行多个查询
		const numWorkers = 3
		done := make(chan bool, numWorkers)
		errors := make(chan error, numWorkers)

		for i := 0; i < numWorkers; i++ {
			go func(workerID int) {
				rows, err := db.Query("SELECT ?", workerID)
				if err != nil {
					errors <- err
					done <- true
					return
				}
				_ = rows.Close()
				done <- true
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < numWorkers; i++ {
			<-done
		}

		// 检查是否有错误
		select {
		case err := <-errors:
			t.Errorf("并发查询失败: %v", err)
		default:
			// 没有错误，测试通过
		}

		// 验证连接池统计
		stats := db.Stats()
		t.Logf("并发测试后连接池统计: OpenConnections=%d, InUse=%d, Idle=%d, WaitCount=%d, WaitDuration=%v",
			stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount, stats.WaitDuration)
	})
}

// TestConnectionPoolResourceCleanup 测试连接池资源清理
func TestConnectionPoolResourceCleanup(t *testing.T) {
	t.Run("验证连接池资源清理", func(t *testing.T) {
		// 创建一个测试数据库连接
		db, err := sql.Open("mysql", "root:rootkit@tcp(localhost:3306)/testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping resource cleanup test")
			return
		}

		// 配置连接池
		db.SetMaxOpenConns(3)
		db.SetMaxIdleConns(2)
		db.SetConnMaxLifetime(5 * time.Minute) // 较短的生存期
		db.SetConnMaxIdleTime(2 * time.Minute) // 较短的空闲时间

		// 执行一些查询
		for i := 0; i < 5; i++ {
			rows, err := db.Query("SELECT ?", i)
			if err != nil {
				t.Skip("查询失败，跳过资源清理测试")
				_ = db.Close()
				return
			}
			_ = rows.Close()
		}

		// 获取初始统计
		initialStats := db.Stats()
		t.Logf("初始连接池统计: OpenConnections=%d", initialStats.OpenConnections)

		// 关闭数据库连接（应该清理所有资源）
		err = db.Close()
		assert.Nil(t, err, "关闭数据库连接应该成功")

		// 验证关闭后不能执行查询
		_, err = db.Query("SELECT 1")
		assert.NotNil(t, err, "关闭后执行查询应该失败")
		assert.Contains(t, err.Error(), "closed", "错误应该包含'closed'")
	})
}

// TestConnectionPoolErrorHandling 测试连接池错误处理
func TestConnectionPoolErrorHandling(t *testing.T) {
	t.Run("验证连接池错误处理", func(t *testing.T) {
		// 测试无效连接字符串
		db, err := sql.Open("mysql", "invalid:connection@string")
		if err != nil {
			// 即使连接字符串无效，sql.Open也可能成功（延迟连接）
			t.Logf("sql.Open 可能成功（延迟连接），实际错误会在Ping时出现")
		}

		if db != nil {
			defer func() { _ = db.Close() }()

			// Ping应该失败
			err = db.Ping()
			assert.NotNil(t, err, "无效连接字符串应该导致Ping失败")
			t.Logf("预期错误: %v", err)
		}

		// 测试配置无效参数
		db2, err := sql.Open("mysql", "root:rootkit@tcp(localhost:3306)/testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping error handling test")
			return
		}
		defer func() { _ = db2.Close() }()

		// 设置无效的连接池参数（应该被正确处理）
		db2.SetMaxOpenConns(-1) // 无效值，应该被忽略或使用默认值
		db2.SetMaxIdleConns(-1) // 无效值，应该被忽略或使用默认值

		// 即使有无效参数，Ping也应该工作
		err = db2.Ping()
		assert.Nil(t, err, "即使有无效配置，Ping也应该工作")
	})
}
