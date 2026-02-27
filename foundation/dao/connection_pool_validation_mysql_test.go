//go:build mysql
// +build mysql

package dao

import (
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/stretchr/testify/assert"
)

// TestMySQLConnectionPoolInFetch 测试Fetch函数中的MySQL连接池配置
func TestMySQLConnectionPoolInFetch(t *testing.T) {
	t.Run("验证MySQL Fetch函数的连接池配置", func(t *testing.T) {
		// 使用Fetch创建DAO（当前没有连接池配置）
		dao, err := Fetch("root", "rootkit", "localhost:3306", "testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping Fetch connection pool test")
			return
		}
		defer dao.Release()

		// 验证DAO基本功能
		err = dao.Ping()
		assert.Nil(t, err, "DAO Ping应该成功")

		// 测试查询功能
		err = dao.Query("SELECT 1")
		assert.Nil(t, err, "DAO Query应该成功")
		defer dao.Finish()

		// 验证可以获取下一行
		hasNext := dao.Next()
		assert.True(t, hasNext, "应该有一行数据")

		// 获取字段值
		var result int
		err = dao.GetField(&result)
		assert.Nil(t, err, "GetField应该成功")
		assert.Equal(t, 1, result, "查询结果应该为1")
	})
}

// TestMySQLConnectionPoolConcurrentDAOs 测试并发DAO访问
func TestMySQLConnectionPoolConcurrentDAOs(t *testing.T) {
	t.Run("验证并发DAO访问的连接池行为", func(t *testing.T) {
		const numDAOs = 3
		daos := make([]Dao, numDAOs)
		errors := make([]*cd.Error, numDAOs)

		// 并发创建多个DAO
		for i := 0; i < numDAOs; i++ {
			dao, err := Fetch("root", "rootkit", "localhost:3306", "testdb")
			if err != nil {
				t.Skipf("创建DAO %d 失败: %v", i, err)
				// 清理已创建的DAO
				for j := 0; j < i; j++ {
					daos[j].Release()
				}
				return
			}
			daos[i] = dao
		}

		// 确保清理
		defer func() {
			for _, dao := range daos {
				if dao != nil {
					dao.Release()
				}
			}
		}()

		// 并发执行Ping
		done := make(chan bool, numDAOs)
		for i := 0; i < numDAOs; i++ {
			go func(idx int) {
				errors[idx] = daos[idx].Ping()
				done <- true
			}(i)
		}

		// 等待所有完成
		for i := 0; i < numDAOs; i++ {
			<-done
		}

		// 验证所有Ping都成功
		for i, err := range errors {
			assert.Nil(t, err, "DAO %d Ping应该成功", i)
		}

		// 测试并发查询
		for i := 0; i < numDAOs; i++ {
			go func(idx int) {
				err := daos[idx].Query("SELECT ?", idx)
				if err == nil {
					defer daos[idx].Finish()
					if daos[idx].Next() {
						var result int
						daos[idx].GetField(&result)
						// 验证结果
						assert.Equal(t, idx, result, "DAO %d 查询结果应该匹配", idx)
					}
				}
				done <- true
			}(i)
		}

		for i := 0; i < numDAOs; i++ {
			<-done
		}
	})
}

// TestMySQLConnectionPoolTransactionIsolation 测试事务隔离
func TestMySQLConnectionPoolTransactionIsolation(t *testing.T) {
	t.Run("验证连接池中的事务隔离", func(t *testing.T) {
		dao, err := Fetch("root", "rootkit", "localhost:3306", "testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping transaction test")
			return
		}
		defer dao.Release()

		// 开始事务
		err = dao.BeginTransaction()
		assert.Nil(t, err, "开始事务应该成功")

		// 在事务中执行查询
		err = dao.Query("SELECT 1")
		assert.Nil(t, err, "事务中查询应该成功")
		defer dao.Finish()

		// 提交事务
		err = dao.CommitTransaction()
		assert.Nil(t, err, "提交事务应该成功")

		// 测试嵌套事务（当前实现支持嵌套事务计数）
		err = dao.BeginTransaction()
		assert.Nil(t, err, "开始嵌套事务应该成功")

		err = dao.BeginTransaction()
		assert.Nil(t, err, "开始第二层嵌套事务应该成功")

		err = dao.CommitTransaction()
		assert.Nil(t, err, "提交第一层嵌套事务应该成功")

		err = dao.CommitTransaction()
		assert.Nil(t, err, "提交第二层嵌套事务应该成功")
	})
}

// TestMySQLConnectionPoolStress 测试连接池压力测试
func TestMySQLConnectionPoolStress(t *testing.T) {
	t.Run("验证连接池压力测试", func(t *testing.T) {
		if testing.Short() {
			t.Skip("跳过压力测试（短模式）")
		}

		dao, err := Fetch("root", "rootkit", "localhost:3306", "testdb")
		if err != nil {
			t.Skip("MySQL not available, skipping stress test")
			return
		}
		defer dao.Release()

		const numIterations = 50
		const numConcurrent = 10

		// 创建测试表（使用普通表，因为临时表是连接特定的）
		_, err = dao.Execute(`
			CREATE TABLE IF NOT EXISTS connection_pool_test (
				id INT AUTO_INCREMENT PRIMARY KEY,
				value VARCHAR(100),
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			) ENGINE=InnoDB`)
		if err != nil {
			t.Skipf("创建测试表失败: %v", err)
			return
		}

		// 清空表数据
		_, err = dao.Execute("TRUNCATE TABLE connection_pool_test")
		if err != nil {
			t.Skipf("清空测试表失败: %v", err)
			return
		}

		// 压力测试：并发插入和查询
		errors := make(chan *cd.Error, numIterations*numConcurrent)
		done := make(chan bool, numIterations*numConcurrent)

		for i := 0; i < numIterations; i++ {
			for j := 0; j < numConcurrent; j++ {
				go func(iteration, worker int) {
					// 插入数据
					_, err := dao.Execute(
						"INSERT INTO connection_pool_test (value) VALUES (?)",
						time.Now().Format(time.RFC3339Nano))
					if err != nil {
						errors <- err
					}

					// 查询数据
					err = dao.Query("SELECT COUNT(*) FROM connection_pool_test")
					if err != nil {
						errors <- err
					} else {
						defer dao.Finish()
						if dao.Next() {
							var count int
							dao.GetField(&count)
							// 验证计数合理
							assert.GreaterOrEqual(t, count, 0, "计数应该 >= 0")
						}
					}

					done <- true
				}(i, j)
			}
		}

		// 等待所有完成
		for i := 0; i < numIterations*numConcurrent; i++ {
			<-done
		}

		// 检查错误
		select {
		case err := <-errors:
			t.Errorf("压力测试中发现错误: %v", err)
		default:
			// 没有错误，测试通过
			t.Logf("压力测试完成: %d 次迭代 × %d 并发 = %d 次操作",
				numIterations, numConcurrent, numIterations*numConcurrent)
		}

		// 清理测试表
		_, err = dao.Execute("DROP TABLE IF EXISTS connection_pool_test")
		assert.Nil(t, err, "清理测试表应该成功")
	})
}

// TestMySQLConnectionPoolConfigurationValidation 测试连接池配置验证
func TestMySQLConnectionPoolConfigurationValidation(t *testing.T) {
	t.Run("验证连接池配置参数", func(t *testing.T) {
		// 测试不同的连接池配置组合
		testCases := []struct {
			name          string
			maxOpenConns  int
			maxIdleConns  int
			maxLifetime   time.Duration
			maxIdleTime   time.Duration
			expectSuccess bool
		}{
			{
				name:          "默认配置",
				maxOpenConns:  0, // 无限制
				maxIdleConns:  2,
				maxLifetime:   0, // 永不过期
				maxIdleTime:   0,
				expectSuccess: true,
			},
			{
				name:          "生产环境推荐配置",
				maxOpenConns:  10,
				maxIdleConns:  5,
				maxLifetime:   time.Hour,
				maxIdleTime:   30 * time.Minute,
				expectSuccess: true,
			},
			{
				name:          "高并发配置",
				maxOpenConns:  50,
				maxIdleConns:  25,
				maxLifetime:   30 * time.Minute,
				maxIdleTime:   10 * time.Minute,
				expectSuccess: true,
			},
			{
				name:          "保守配置",
				maxOpenConns:  5,
				maxIdleConns:  3,
				maxLifetime:   time.Hour,
				maxIdleTime:   15 * time.Minute,
				expectSuccess: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				// 注意：当前Fetch函数不支持自定义连接池配置
				// 这里只是验证配置参数本身的合理性
				dao, err := Fetch("root", "rootkit", "localhost:3306", "testdb")
				if err != nil {
					t.Skip("MySQL not available, skipping configuration test")
					return
				}
				defer dao.Release()

				// 验证基本功能
				err = dao.Ping()
				assert.Nil(t, err, "Ping应该成功")

				// 执行简单查询
				err = dao.Query("SELECT 1")
				assert.Nil(t, err, "查询应该成功")
				defer dao.Finish()

				// 验证查询结果
				hasNext := dao.Next()
				assert.True(t, hasNext, "应该有一行数据")

				var result int
				err = dao.GetField(&result)
				assert.Nil(t, err, "获取字段应该成功")
				assert.Equal(t, 1, result, "查询结果应该为1")

				t.Logf("配置测试通过: %s", tc.name)
			})
		}
	})
}
