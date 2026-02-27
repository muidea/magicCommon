package dao

import (
	"fmt"
	"runtime"
	"strings"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/util"
)

// tracedError 带堆栈跟踪的错误
type tracedError struct {
	baseError *cd.Error
	stack     []string
}

// newTracedError 创建带堆栈跟踪的错误
func newTracedError(err *cd.Error) *tracedError {
	if err == nil {
		return nil
	}

	// 获取调用堆栈
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:]) // 跳过 runtime.Callers, newTracedError, 和调用者
	frames := runtime.CallersFrames(pcs[:n])

	var stack []string
	for {
		frame, more := frames.Next()
		// 过滤掉 runtime 和 testing 相关的帧
		if !strings.Contains(frame.File, "runtime/") &&
			!strings.Contains(frame.File, "testing/") &&
			!strings.Contains(frame.Function, "runtime.") &&
			!strings.Contains(frame.Function, "testing.") {
			stack = append(stack, fmt.Sprintf("%s\n\t%s:%d", frame.Function, frame.File, frame.Line))
		}
		if !more {
			break
		}
	}

	return &tracedError{
		baseError: err,
		stack:     stack,
	}
}

// Error 实现 error 接口
func (e *tracedError) Error() string {
	if e == nil || e.baseError == nil {
		return ""
	}

	if len(e.stack) == 0 {
		return e.baseError.Error()
	}

	return fmt.Sprintf("%s\nStack Trace:\n%s", e.baseError.Error(), strings.Join(e.stack, "\n"))
}

// StackTrace 返回堆栈跟踪
func (e *tracedError) StackTrace() string {
	if e == nil || len(e.stack) == 0 {
		return ""
	}
	return strings.Join(e.stack, "\n")
}

// WrapErrorWithTrace 包装错误并添加堆栈跟踪
func WrapErrorWithTrace(err error) *cd.Error {
	if err == nil {
		return nil
	}

	// 如果已经是带堆栈跟踪的错误，直接返回
	if cdErr, ok := err.(*cd.Error); ok && IsTracedError(cdErr) {
		return cdErr
	}

	// 如果已经是 *cd.Error，包装它
	var cdErr *cd.Error
	if e, ok := err.(*cd.Error); ok {
		cdErr = e
	} else {
		cdErr = util.DatabaseErrorFactory.Wrap(cd.DatabaseError, err, "database operation with trace")
	}

	// 创建带堆栈跟踪的错误
	tracedErr := newTracedError(cdErr)
	return &cd.Error{
		Code:    cdErr.Code,
		Message: tracedErr.Error(), // 包含堆栈跟踪的消息
	}
}

// GetOriginalError 获取原始错误（不带堆栈跟踪）
func GetOriginalError(err *cd.Error) *cd.Error {
	if err == nil {
		return nil
	}
	// 简单实现：返回相同的错误
	// 在实际实现中，可能需要解析消息以提取原始错误
	return err
}

// IsTracedError 检查是否为带堆栈跟踪的错误
func IsTracedError(err *cd.Error) bool {
	if err == nil {
		return false
	}
	// 检查消息是否包含堆栈跟踪
	return strings.Contains(err.Message, "Stack Trace:")
}

// ExtractStackTrace 从错误中提取堆栈跟踪
func ExtractStackTrace(err *cd.Error) string {
	if err == nil {
		return ""
	}

	if !IsTracedError(err) {
		return ""
	}

	// 从消息中提取堆栈跟踪
	parts := strings.Split(err.Message, "Stack Trace:\n")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
