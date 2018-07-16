package def

import "muidea.com/magicCommon/model"

// LoginAccountParam 账号登陆参数
type LoginAccountParam struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// LoginAccountResult 账号登陆结果
type LoginAccountResult struct {
	Result
	OnlineUser model.AccountOnlineView `json:"onlineUser"`
	SessionID  string                  `json:"sessionID"`
	AuthToken  string                  `json:"authToken"`
}

// LogoutAccountResult 账号登出结果
type LogoutAccountResult Result

// StatusAccountResult 获取账号状态结果
type StatusAccountResult struct {
	Result
	OnlineUser model.AccountOnlineView `json:"onlineUser"`
	SessionID  string                  `json:"sessionID"`
}
