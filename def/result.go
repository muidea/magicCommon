package def

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
	ErrorCode int    `json:"errorCode"`
	Reason    string `json:"reason"`
}

// Success 成功
func (result *Result) Success() bool {
	return result.ErrorCode == Success
}

// Fail 失败
func (result *Result) Fail() bool {
	return result.ErrorCode != Success
}
