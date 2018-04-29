package model

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id"`
	// Name 名称
	Name string `json:"name"`
}

// Status 状态
type Status Unit

// ModuleType 模块类型
type ModuleType Unit
