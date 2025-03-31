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
	// UnExpected 意外错误
	UnExpected
	// Duplicated 重复
	Duplicated
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

// Result 处理结果
// Code 错误码
// Reason 错误信息
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
		Error: NewError(UnExpected, "unexpected error"),
	}
}
