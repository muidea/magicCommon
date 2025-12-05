package def

import (
	"fmt"
)

type Code int

const (
	Success = iota
	// UnKnownError 未知错误
	UnKnownError
	// NotFound 未找到
	NotFound
	// InvalidParam 无效参数
	InvalidParameter
	// IllegalParam 非法参数
	IllegalParam
	// InvalidAuthority 非法授权
	InvalidAuthority
	// Unexpected 意外错误
	Unexpected
	// Duplicated 重复
	Duplicated
	// DatabaseError 数据库错误
	DatabaseError
	// Timeout 超时
	Timeout
	// NetworkError 网络错误
	NetworkError
	// Unauthorized 未授权
	Unauthorized
	// Forbidden 禁止访问
	Forbidden
	// ResourceExhausted 资源耗尽
	ResourceExhausted
	// TooManyRequests 请求过多
	TooManyRequests
	// ServiceUnavailable 服务不可用
	ServiceUnavailable
	// NotImplemented 未实现
	NotImplemented
	// BadGateway 网关错误
	BadGateway
	// DataCorrupted 数据损坏
	DataCorrupted
	// VersionConflict 版本冲突
	VersionConflict
	// ExternalServiceError 外部服务错误
	ExternalServiceError
	// InvalidOperation 无效操作
	InvalidOperation
)

type Error struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return fmt.Sprintf("code:%d, message:%s", e.Code, e.Message)
}

func NewError(errorCode Code, errorMessage string) *Error {
	return &Error{
		Code:    errorCode,
		Message: errorMessage,
	}
}

// ResultWithError 表示包含错误的结果接口
type ResultWithError interface {
	GetError() *Error
}

type Result struct {
	Error *Error `json:"error"`
}

// Success 成功
func (s *Result) Success() bool {
	return s.Error == nil || s.Error.Code == Success
}

// Fail 失败
func (s *Result) Fail() bool {
	return s.Error != nil && s.Error.Code != Success
}

func NewResult() *Result {
	return &Result{
		Error: NewError(Unexpected, "unexpected error"),
	}
}

// GetError 返回错误指针，便于实现 ResultWithError 接口
func (r *Result) GetError() *Error {
	return r.Error
}

// SetError 设置错误
func (r *Result) SetError(err *Error) {
	r.Error = err
}

// NewSuccessResult 创建一个成功的 Result（无错误）
func NewSuccessResult() *Result {
	return &Result{
		Error: nil,
	}
}
