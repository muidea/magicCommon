package def

import "muidea.com/magicCommon/model"

// GetACLListResult 查询ACL结果
type GetACLListResult struct {
	Result
	ACL []model.ACLView `json:"acl"`
}

// GetACLResult 查询指定ACL结果
type GetACLResult struct {
	Result
	ACL model.ACLView `json:"acl"`
}

// CreateACLParam 新建ACL参数
type CreateACLParam struct {
	URL       string `json:"url"`
	Method    string `json:"method"`
	Module    string `json:"module"`
	AuthGroup int    `json:"authGroup"`
}

// CreateACLResult 新建ACL结果
type CreateACLResult struct {
	Result
	ACL model.ACLView `json:"acl"`
}

// DestroyACLResult 删除ACL结果
type DestroyACLResult Result

// UpdateACLParam 更新ACL参数
type UpdateACLParam struct {
	URL       string `json:"url"`
	Method    string `json:"method"`
	Module    string `json:"module"`
	AuthGroup int    `json:"authGroup"`
	Status    int    `json:"status"`
}

// UpdateACLResult 更新ACL结果
type UpdateACLResult Result

// UpdateACLStatusParam 更新ACL状态参数
type UpdateACLStatusParam struct {
	EnableList  []int `json:"enablelist"`
	DisableList []int `json:"disablelist"`
}

// UpdateACLStatusResult 更新ACL状态结果
type UpdateACLStatusResult Result

// GetAuthGroupResult 查询AuthGroup结果
type GetAuthGroupResult struct {
	Result
	AuthGroup model.AuthGroup `json:"authGroup"`
}

// UpdateAuthGroupParam 更新AuthGroup请求
type UpdateAuthGroupParam struct {
	AuthGroup int `json:"authGroup"`
}

// UpdateAuthGroupResult 更新AuthGroup结果
type UpdateAuthGroupResult Result

// QueryEndpointListResult 查询EndpointList结果
type QueryEndpointListResult struct {
	Result
	Endpoint []model.EndpointView `json:"endpoint"`
}

// QueryEndpointResult 查询Endpoint结果
type QueryEndpointResult struct {
	Result
	Endpoint model.EndpointView `json:"endpoint"`
}

// CreateEndpointParam 新建Endpoint参数
type CreateEndpointParam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	User        []int  `json:"user"`
	Status      int    `json:"status"`
}

// CreateEndpointResult 新建Endpoint结果
type CreateEndpointResult struct {
	Result
	Endpoint model.EndpointView `json:"endpoint"`
}

// DestroyEndpointResult 删除Endpoint结果
type DestroyEndpointResult Result

// UpdateEndpointParam 更新Endpoint参数
type UpdateEndpointParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	User        []int  `json:"user"`
	Status      int    `json:"status"`
}

// UpdateEndpointResult 更新Endpoint结果
type UpdateEndpointResult struct {
	Result
	Endpoint model.EndpointView `json:"endpoint"`
}

// VerifyEndpointResult 验证Endpoint结果
type VerifyEndpointResult struct {
	Result
	SessionID string `json:"sessionID"`
}

// GetModuleUserInfoListResult 查询ModuleUserInfoList结果
type GetModuleUserInfoListResult struct {
	Result
	Module []model.ModuleUserInfoView `json:"module"`
}

// GetModuleAuthGroupInfoResult 查询Module结果
type GetModuleAuthGroupInfoResult struct {
	Result
	Module model.ModuleUserAuthGroupView `json:"module"`
}

// UpdateUserAuthGroupParam 更新Module用户的AuthGroup参数
type UpdateUserAuthGroupParam struct {
	UserAuthGroup []model.UserAuthGroup `json:"userAuthGroup"`
}

// UpdateUserAuthGroupResult 更新Module用户的AuthGroup结果
type UpdateUserAuthGroupResult Result

// GetUserModuleInfoListResult 查询UserModuleInfoList 结果
type GetUserModuleInfoListResult struct {
	Result
	User []model.UserModuleInfoView `json:"user"`
}

// GetUserAuthGroupInfoResult 查询指定用户对用Module的AuthGroup结果
type GetUserAuthGroupInfoResult struct {
	Result
	User model.UserModuleAuthGroupView `json:"user"`
}

// UpdateModuleAuthGroupParam 更新用户对应Module的AuthGroup参数
type UpdateModuleAuthGroupParam struct {
	ModuleAuthGroup []model.ModuleAuthGroup `json:"moduleAuthGroup"`
}

// UpdateModuleAuthGroupResult 更新用户对应Module的AuthGroup结果
type UpdateModuleAuthGroupResult Result