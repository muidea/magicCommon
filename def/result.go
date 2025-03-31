package def

import (
	"encoding/json"
	"fmt"
)

type Code int

const (
	// UnKnownError 未知错误
	UnKnownError = iota
	// NotFound 未找到
	NotFound
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

type Result struct {
	Error *Error `json:"error"`
}

type CommonResult struct {
	Result
	Value json.RawMessage `json:"value"`
}

type CommonSliceResult struct {
	Result
	Total int64           `json:"total"`
	Value json.RawMessage `json:"values"`
}

// Success 成功
func (s *Result) Success() bool {
	return s.Error == nil || s.Error.Code == 0
}

// Fail 失败
func (s *Result) Fail() bool {
	return s.Error != nil && s.Error.Code != 0
}

func NewResult() Result {
	return Result{
		Error: NewError(Unexpected, "unexpected error"),
	}
}
