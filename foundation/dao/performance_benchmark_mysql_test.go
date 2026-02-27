//go:build mysql
// +build mysql

package dao

import (
	"testing"
)

// BenchmarkCriticalOperations 关键操作性能基准测试
// 目的：建立关键操作的性能基准，用于后续性能回归测试

// BenchmarkPerf_BaseDao_String 测试 String 方法性能
func BenchmarkPerf_BaseDao_String(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:3306",
		dbName:  "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseDao.String()
	}
}

// BenchmarkPerf_BaseDao_DBName 测试 DBName 方法性能
func BenchmarkPerf_BaseDao_DBName(b *testing.B) {
	baseDao := &BaseDao{
		dbName: "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseDao.DBName()
	}
}

// BenchmarkPerf_WrapError_Nil 测试包装 nil 错误性能
func BenchmarkPerf_WrapError_Nil(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapError(nil)
	}
}

// BenchmarkPerf_WrapError_StandardError 测试包装标准错误性能
func BenchmarkPerf_WrapError_StandardError(b *testing.B) {
	err := ErrDatabaseNotInitialized

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapError(err)
	}
}

// BenchmarkPerf_ErrorComparison 测试错误比较性能
func BenchmarkPerf_ErrorComparison(b *testing.B) {
	err1 := ErrDatabaseNotInitialized
	err2 := ErrTransactionActive

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err1.Code == err2.Code
	}
}

// BenchmarkPerf_ConcurrentOperations 并发操作性能测试
func BenchmarkPerf_ConcurrentOperations(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:3306",
		dbName:  "testdb",
	}

	b.Run("String", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = baseDao.String()
			}
		})
	})

	b.Run("DBName", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = baseDao.DBName()
			}
		})
	})
}

// BenchmarkPerf_ErrorHandling 错误处理性能测试
func BenchmarkPerf_ErrorHandling(b *testing.B) {
	baseDao := &BaseDao{dbHandle: nil}

	b.Run("Ping_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.Ping()
		}
	})

	b.Run("BeginTransaction_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.BeginTransaction()
		}
	})

	b.Run("Query_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.Query("SELECT 1")
		}
	})

	b.Run("Execute_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Execute("SELECT 1")
		}
	})
}

// BenchmarkPerf_ResultSetMethods 结果集方法性能测试
func BenchmarkPerf_ResultSetMethods(b *testing.B) {
	baseDao := &BaseDao{rowsHandle: nil}

	b.Run("Next_NoResultSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.Next()
		}
	})

	b.Run("Finish_NoResultSet", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.Finish()
		}
	})

	b.Run("GetField_Error", func(b *testing.B) {
		var value int
		for i := 0; i < b.N; i++ {
			_ = baseDao.GetField(&value)
		}
	})
}

// BenchmarkPerf_TransactionMethods 事务方法性能测试
func BenchmarkPerf_TransactionMethods(b *testing.B) {
	baseDao := &BaseDao{}

	b.Run("BeginTransaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.BeginTransaction()
		}
	})

	b.Run("CommitTransaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.CommitTransaction()
		}
	})

	b.Run("RollbackTransaction", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.RollbackTransaction()
		}
	})
}

// BenchmarkPerf_CRUDMethods CRUD方法性能测试
func BenchmarkPerf_CRUDMethods(b *testing.B) {
	baseDao := &BaseDao{}

	b.Run("Insert_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Insert("INSERT INTO test VALUES (?)", i)
		}
	})

	b.Run("Update_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Update("UPDATE test SET value = ? WHERE id = ?", i, i)
		}
	})

	b.Run("Delete_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Delete("DELETE FROM test WHERE id = ?", i)
		}
	})

	b.Run("Execute_Error", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Execute("SELECT * FROM test WHERE id = ?", i)
		}
	})
}

// BenchmarkPerf_MemoryAllocation 内存分配性能测试
func BenchmarkPerf_MemoryAllocation(b *testing.B) {
	b.Run("NewBaseDao_Struct", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = &BaseDao{
				address:  "localhost:3306",
				dbName:   "testdb",
				user:     "user",
				password: "pass",
			}
		}
	})

	b.Run("Error_Creation", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ErrDatabaseNotInitialized
		}
	})
}

// BenchmarkPerf_CompositeOperations 复合操作性能测试
func BenchmarkPerf_CompositeOperations(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:3306",
		dbName:  "testdb",
	}

	b.Run("String_DBName_Combo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.String()
			_ = baseDao.DBName()
		}
	})

	b.Run("Error_Handling_Combo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = baseDao.Ping()
			_ = baseDao.BeginTransaction()
			_ = baseDao.CommitTransaction()
		}
	})
}

// BenchmarkPerf_LongRunning 长时间运行性能测试
func BenchmarkPerf_LongRunning(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:3306",
		dbName:  "testdb",
	}

	// 模拟长时间运行场景
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 混合各种操作
		_ = baseDao.String()
		_ = baseDao.DBName()
		_ = baseDao.Ping()
		_ = baseDao.BeginTransaction()
		_ = baseDao.CommitTransaction()
		_ = baseDao.Query("SELECT 1")
		_, _ = baseDao.Execute("SELECT 1")
		_ = baseDao.Finish()
	}
}
