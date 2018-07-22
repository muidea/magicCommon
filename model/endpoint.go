package model

// Endpoint 终端
type Endpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	User        []int  `json:"user"`
	Status      int    `json:"status"`
	AuthToken   string `json:"authToken"`
}

// EndpointView 终端视图
type EndpointView struct {
	Endpoint
	User []User `json:"user"`
}

// EndpointSummary endpoint管理摘要
type EndpointSummary []UnitSummary
