package model

// Syslog 系统日志
type Syslog struct {
	ID        int    `json:"id"`
	User      string `json:"user"`
	Operation string `json:"operation"`
	DateTime  string `json:"dateTime"`
	Source    string `json:"source"`
}
