package model

// AccountOnlineView 在线用户信息
type AccountOnlineView struct {
	User
	LoginTime  int64  `json:"loginTime"`  // 登陆时间
	UpdateTime int64  `json:"updateTime"` // 更新时间
	Address    string `json:"address"`    // 访问IP
	AuthToken  string `json:"authToken"`
}
