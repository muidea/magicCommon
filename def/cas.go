package def

import "github.com/muidea/magicCommon/model"

// LoginAccountParam 账号登陆参数
type LoginAccountParam struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// LoginAccountResult 账号登陆结果
type LoginAccountResult struct {
	Result
	OnlineEntry model.OnlineEntryView `json:"onlineEntry"`
	SessionID   string                `json:"sessionID"`
	AuthToken   string                `json:"authToken"`
}

// LogoutAccountResult 账号登出结果
type LogoutAccountResult Result

// StatusAccountResult 获取账号状态结果
type StatusAccountResult LoginAccountResult

// ChangeAccountPasswordParam 更改密码参数
type ChangeAccountPasswordParam struct {
	OldPassword string `json:"oldPassword"`
	NewPassword string `json:"newPassword"`
}

// ChangeAccountPasswordResult 更改密码结果
type ChangeAccountPasswordResult Result

// LoginEndpointParam Endpoint登陆请求
type LoginEndpointParam struct {
	IdentifyID string `json:"identifyID"`
	AuthToken  string `json:"authToken"`
}

// LoginEndpointResult Endpoint登陆结果
type LoginEndpointResult struct {
	Result
	SessionID string `json:"sessionID"`
	AuthToken string `json:"authToken"`
}

// LogoutEndpointResult Endpoint登出结果
type LogoutEndpointResult Result

// StatusEndpointResult 获取Endpoint状态结果
type StatusEndpointResult LoginEndpointResult
