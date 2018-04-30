package model

// AuthGroup 授权组
type AuthGroup struct {
	Unit
	Description string `json:"description"`
}

// ACL 访问控制
type ACL struct {
	ID     int    `json:"id"`
	URL    string `json:"url"`
	Method string `json:"method"`
	Status int    `json:"status"`
}

// ACLView acl
type ACLView struct {
	ACL
	Status Status `json:"status"`
}

// ACLDetail 访问控制列表
type ACLDetail struct {
	ACL
	Module    string `json:"module"`
	AuthGroup int    `json:"authGroup"`
}

// ACLDetailView ACL显示信息
type ACLDetailView struct {
	ACLDetail
	Status    Status    `json:"status"`
	Module    Module    `json:"module"`
	AuthGroup AuthGroup `json:"authGroup"`
}

// ModuleAccountInfo 模块的用户信息
type ModuleAccountInfo struct {
	Module  string `json:"module"`
	Account []int  `json:"account"`
}

// ModuleAccountInfoView 模块的用户信息
type ModuleAccountInfoView struct {
	Module
	Account []Account `json:"account"`
}

// AccountAuthGroup 用户授权组
type AccountAuthGroup struct {
	Account   int `json:"account"`
	AuthGroup int `json:"authGroup"`
}

// AccountAuthGroupView 用户授权组显示信息
type AccountAuthGroupView struct {
	Account
	AuthGroup AuthGroup `json:"authGroup"`
}

// ModuleAccountAuthGroupView 模块的用户授权组显示信息
type ModuleAccountAuthGroupView struct {
	ModuleDetail
	Status           Status
	AccountAuthGroup []AccountAuthGroupView `json:"accountAuthGroup"`
}

// AccountModuleInfo 用户模块信息
type AccountModuleInfo struct {
	Account int      `json:"account"`
	Module  []string `json:"module"`
}

// AccountModuleInfoView 用户模块信息
type AccountModuleInfoView struct {
	Account
	Module []Module `json:"module"`
}

// ModuleAuthGroup 模块授权组
type ModuleAuthGroup struct {
	Module    string `json:"module"`
	AuthGroup int    `json:"authGroup"`
}

// ModuleAuthGroupView 模块授权组显示信息
type ModuleAuthGroupView struct {
	Module
	AuthGroup AuthGroup `json:"authGroup"`
}

// AccountModuleAuthGroupView 用户的模块授权组显示信息
type AccountModuleAuthGroupView struct {
	AccountDetailView
	ModuleAuthGroup []ModuleAuthGroupView `json:"moduleAuthGroup"`
}
