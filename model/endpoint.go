package model

// Endpoint 终端
type Endpoint struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	User        []int  `json:"user"`
	Status      int    `json:"status"`
	AccessToken string `json:"accessToken"`
}

// EndpointView 终端视图
type EndpointView struct {
	Endpoint
	User []User `json:"user"`
}
