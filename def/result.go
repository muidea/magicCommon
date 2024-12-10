package def

import (
	"encoding/json"
	"fmt"
)

type ErrorCode int

const (
	// Succeeded 成功
	Succeeded = iota
	// Warned 警告
	Warned = 200000
	// NoExist 对象不存在
	NoExist = 200001
	// Failed 失败
	Failed = 500000
	// IllegalParam 非法参数
	IllegalParam = 500001
	// InvalidAuthority 非法授权
	InvalidAuthority = 500002
	// Redirect 对象转移
	Redirect = 500003
	// UnExpected 意外错误
	UnExpected = 500004
	// Duplicated 重复
	Duplicated = 500005
)

// Result 处理结果
// ErrorCode 错误码
// Reason 错误信息
type Result struct {
	ErrorCode ErrorCode `json:"errorCode"`
	Reason    string    `json:"reason"`
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
	return s.ErrorCode == Succeeded
}

func (s *Result) Warn() bool {
	return s.ErrorCode >= Warned && s.ErrorCode < Failed
}

// Fail 失败
func (s *Result) Fail() bool {
	return s.ErrorCode != Succeeded
}

func (s *Result) Error() string {
	if s.ErrorCode == Succeeded {
		return ""
	}

	if s.Reason != "" {
		return fmt.Sprintf("errorCode:%v, reason:%v", s.ErrorCode, s.Reason)
	}

	return fmt.Sprintf("errorCode:%v", s.ErrorCode)
}

func (s *Result) String() string {
	return fmt.Sprintf("errorCode:%v, reason:%v", s.ErrorCode, s.Reason)
}

func NewError(errCode ErrorCode, reason string) *Result {
	return &Result{ErrorCode: errCode, Reason: reason}
}
