//go:build !mysql
// +build !mysql

package dao

import (
	"testing"
)

// BenchmarkBaseDao_String 测试 String 方法的性能
func BenchmarkBaseDao_String(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:5432",
		dbName:  "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseDao.String()
	}
}

// BenchmarkBaseDao_DBName 测试 DBName 方法的性能
func BenchmarkBaseDao_DBName(b *testing.B) {
	baseDao := &BaseDao{
		dbName: "testdb",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = baseDao.DBName()
	}
}

// BenchmarkWrapError 测试错误包装的性能
func BenchmarkWrapError(b *testing.B) {
	err := error(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapError(err)
	}
}

// BenchmarkWrapError_WithError 测试有错误时的包装性能
func BenchmarkWrapError_WithError(b *testing.B) {
	err := ErrDatabaseNotInitialized

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapError(err)
	}
}

// BenchmarkBaseDao_ConcurrentString 并发测试 String 方法
func BenchmarkBaseDao_ConcurrentString(b *testing.B) {
	baseDao := &BaseDao{
		address: "localhost:5432",
		dbName:  "testdb",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = baseDao.String()
		}
	})
}

// BenchmarkBaseDao_ConcurrentDBName 并发测试 DBName 方法
func BenchmarkBaseDao_ConcurrentDBName(b *testing.B) {
	baseDao := &BaseDao{
		dbName: "testdb",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = baseDao.DBName()
		}
	})
}

// BenchmarkErrorComparison 测试错误比较性能
func BenchmarkErrorComparison(b *testing.B) {
	err1 := ErrDatabaseNotInitialized
	err2 := ErrTransactionActive

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = err1.Code == err2.Code
	}
}

// BenchmarkNewBaseDao 测试创建 BaseDao 的性能
func BenchmarkNewBaseDao(b *testing.B) {
	// 注意：这个基准测试需要真实的数据库连接
	// 在实际环境中运行
	b.Skip("需要真实的数据库连接，跳过此基准测试")
}

// BenchmarkBaseDao_TransactionMethods 测试事务方法的性能（模拟）
func BenchmarkBaseDao_TransactionMethods(b *testing.B) {
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

// BenchmarkBaseDao_CRUDMethods 测试 CRUD 方法的性能（模拟）
func BenchmarkBaseDao_CRUDMethods(b *testing.B) {
	baseDao := &BaseDao{}

	b.Run("Insert", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Insert("INSERT INTO test VALUES (?)", i)
		}
	})

	b.Run("Update", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Update("UPDATE test SET value = ? WHERE id = ?", i, i)
		}
	})

	b.Run("Delete", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Delete("DELETE FROM test WHERE id = ?", i)
		}
	})

	b.Run("Execute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = baseDao.Execute("SELECT * FROM test WHERE id = ?", i)
		}
	})
}
