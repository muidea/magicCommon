package dao

import (
	"errors"
	"strings"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/stretchr/testify/assert"
)

// TestWrapErrorWithTrace 测试带堆栈跟踪的错误包装
func TestWrapErrorWithTrace(t *testing.T) {
	t.Run("包装 nil 错误", func(t *testing.T) {
		err := WrapErrorWithTrace(nil)
		assert.Nil(t, err)
	})

	t.Run("包装标准错误", func(t *testing.T) {
		stdErr := errors.New("standard error")
		wrappedErr := WrapErrorWithTrace(stdErr)
		assert.NotNil(t, wrappedErr)
		assert.EqualValues(t, cd.DatabaseError, wrappedErr.Code)
		assert.Contains(t, wrappedErr.Message, "standard error")
		assert.Contains(t, wrappedErr.Message, "Stack Trace:")
	})

	t.Run("包装 *cd.Error", func(t *testing.T) {
		cdErr := cd.NewError(cd.InvalidParameter, "invalid param")
		wrappedErr := WrapErrorWithTrace(cdErr)
		assert.NotNil(t, wrappedErr)
		assert.EqualValues(t, cd.InvalidParameter, wrappedErr.Code)
		assert.Contains(t, wrappedErr.Message, "invalid param")
		assert.Contains(t, wrappedErr.Message, "Stack Trace:")
	})

	t.Run("包装带堆栈跟踪的错误", func(t *testing.T) {
		cdErr := cd.NewError(cd.DatabaseError, "database error")
		wrappedErr1 := WrapErrorWithTrace(cdErr)
		wrappedErr2 := WrapErrorWithTrace(wrappedErr1)
		assert.NotNil(t, wrappedErr2)
		assert.Contains(t, wrappedErr2.Message, "database error")
		// 应该只包含一次堆栈跟踪
		stackTraceCount := strings.Count(wrappedErr2.Message, "Stack Trace:")
		assert.Equal(t, 1, stackTraceCount)
	})
}

// TestIsTracedError 测试检查是否为带堆栈跟踪的错误
func TestIsTracedError(t *testing.T) {
	t.Run("nil 错误", func(t *testing.T) {
		assert.False(t, IsTracedError(nil))
	})

	t.Run("普通错误", func(t *testing.T) {
		err := cd.NewError(cd.DatabaseError, "database error")
		assert.False(t, IsTracedError(err))
	})

	t.Run("带堆栈跟踪的错误", func(t *testing.T) {
		stdErr := errors.New("standard error")
		wrappedErr := WrapErrorWithTrace(stdErr)
		assert.True(t, IsTracedError(wrappedErr))
	})

	t.Run("部分匹配的消息", func(t *testing.T) {
		err := cd.NewError(cd.DatabaseError, "Error with Stack Trace: something")
		assert.True(t, IsTracedError(err))
	})
}

// TestExtractStackTrace 测试提取堆栈跟踪
func TestExtractStackTrace(t *testing.T) {
	t.Run("nil 错误", func(t *testing.T) {
		stack := ExtractStackTrace(nil)
		assert.Equal(t, "", stack)
	})

	t.Run("普通错误", func(t *testing.T) {
		err := cd.NewError(cd.DatabaseError, "database error")
		stack := ExtractStackTrace(err)
		assert.Equal(t, "", stack)
	})

	t.Run("带堆栈跟踪的错误", func(t *testing.T) {
		stdErr := errors.New("standard error")
		wrappedErr := WrapErrorWithTrace(stdErr)
		stack := ExtractStackTrace(wrappedErr)
		assert.NotEqual(t, "", stack)
		assert.Contains(t, stack, ".go:")
		assert.NotContains(t, stack, "Stack Trace:")
	})

	t.Run("无效的堆栈跟踪格式", func(t *testing.T) {
		err := cd.NewError(cd.DatabaseError, "Error without proper stack trace")
		stack := ExtractStackTrace(err)
		assert.Equal(t, "", stack)
	})
}

// TestGetOriginalError 测试获取原始错误
func TestGetOriginalError(t *testing.T) {
	t.Run("nil 错误", func(t *testing.T) {
		origErr := GetOriginalError(nil)
		assert.Nil(t, origErr)
	})

	t.Run("普通错误", func(t *testing.T) {
		err := cd.NewError(cd.DatabaseError, "database error")
		origErr := GetOriginalError(err)
		assert.Equal(t, err, origErr)
	})

	t.Run("带堆栈跟踪的错误", func(t *testing.T) {
		stdErr := errors.New("standard error")
		wrappedErr := WrapErrorWithTrace(stdErr)
		origErr := GetOriginalError(wrappedErr)
		assert.NotNil(t, origErr)
		// 注意：当前实现返回相同的错误
		// 在实际实现中，可能需要解析消息
		assert.Equal(t, wrappedErr, origErr)
	})
}

// TestErrorTraceIntegration 测试错误追踪集成
func TestErrorTraceIntegration(t *testing.T) {
	// 模拟一个函数调用链
	func3 := func() *cd.Error {
		// 模拟一个错误
		err := errors.New("simulated error")
		return WrapErrorWithTrace(err)
	}

	func2 := func() *cd.Error {
		return func3()
	}

	func1 := func() *cd.Error {
		return func2()
	}

	err := func1()
	assert.NotNil(t, err)
	assert.True(t, IsTracedError(err))

	stack := ExtractStackTrace(err)
	assert.NotEqual(t, "", stack)

	// 检查堆栈跟踪是否包含相关函数
	// 注意：由于函数是匿名函数，它们可能不会出现在堆栈跟踪中
	// 但堆栈跟踪应该包含这个测试文件
	assert.Contains(t, stack, "error_trace_test.go")
}

// BenchmarkWrapErrorWithTrace 基准测试带堆栈跟踪的错误包装
func BenchmarkWrapErrorWithTrace(b *testing.B) {
	stdErr := errors.New("benchmark error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = WrapErrorWithTrace(stdErr)
	}
}

// BenchmarkIsTracedError 基准测试检查堆栈跟踪
func BenchmarkIsTracedError(b *testing.B) {
	stdErr := errors.New("benchmark error")
	wrappedErr := WrapErrorWithTrace(stdErr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = IsTracedError(wrappedErr)
	}
}

// BenchmarkExtractStackTrace 基准测试提取堆栈跟踪
func BenchmarkExtractStackTrace(b *testing.B) {
	stdErr := errors.New("benchmark error")
	wrappedErr := WrapErrorWithTrace(stdErr)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractStackTrace(wrappedErr)
	}
}
