package model

// OnlineEntryView 在线对象信息
type OnlineEntryView struct {
	User
	LoginTime  int64  `json:"loginTime"`  // 登陆时间
	UpdateTime int64  `json:"updateTime"` // 更新时间
	Address    string `json:"address"`    // 访问IP
	IdentifyID string `json:"identifyID"`
}

// CasSummary Cas摘要
type CasSummary []UnitSummary
