package def

import "fmt"

type ErrorCode int

const (
	// Success 成功
	Success = iota
	// Failed 失败
	Failed
	// IllegalParam 非法参数
	IllegalParam
	// InvalidAuthority 非法授权
	InvalidAuthority
	// NoExist 对象不存在
	NoExist
	// Redirect 对象转移
	Redirect
	// UnExpected 意外错误
	UnExpected
)

// Result 处理结果
// ErrorCode 错误码
// Reason 错误信息
type Result struct {
	ErrorCode ErrorCode `json:"errorCode"`
	Reason    string    `json:"reason"`
}

// Success 成功
func (s *Result) Success() bool {
	return s.ErrorCode == Success
}

// Fail 失败
func (s *Result) Fail() bool {
	return s.ErrorCode != Success
}

func (s *Result) Error() error {
	if s.ErrorCode == Success {
		return nil
	}

	return fmt.Errorf("errorCode:%v, reason:%v", s.ErrorCode, s.Reason)
}

func GetError(errCode ErrorCode, reason string) error {
	if errCode == Success {
		return nil
	}

	return fmt.Errorf("errorCode:%v, reason:%v", errCode, reason)
}
