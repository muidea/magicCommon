package def

import "muidea.com/magicCommon/model"

// QuerySystemConfigResult 查询SystemConfig结果
type QuerySystemConfigResult struct {
	Result
	SystemProperty model.SystemProperty `json:"systemProperty"`
}

// UpdateSystemConfigResult 更新SystemConfig结果
type UpdateSystemConfigResult Result

// QuerySyslogResult 查询Syslog结果
type QuerySyslogResult struct {
	Result
	Total  int            `json:"total"`
	Syslog []model.Syslog `json:"syslog"`
}

// InsertSyslogParam 插入Syslog参数
type InsertSyslogParam struct {
	User      string `json:"user"`
	Operation string `json:"operation"`
	DateTime  string `json:"dateTime"`
	Source    string `json:"source"`
}

// InsertSyslogResult 插入Syslog结果
type InsertSyslogResult Result
