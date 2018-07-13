package def

import (
	"muidea.com/magicCommon/model"
)

// GetGroupResult 获取分组详情
type GetGroupResult struct {
	Result
	Group model.GroupDetailView `json:"group"`
}

// GetGroupListResult 获取分组列表
type GetGroupListResult struct {
	Result
	Group []model.GroupDetailView `json:"group"`
}

// CreateGroupParam 新建分组参数
type CreateGroupParam struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Catalog     model.Group `json:"catalog"`
}

// CreateGroupResult 新建分组结果
type CreateGroupResult struct {
	Result
	Group model.GroupDetailView `json:"group"`
}

// UpdateGroupParam 更新分组参数
type UpdateGroupParam CreateGroupParam

// UpdateGroupResult 更新分组结果
type UpdateGroupResult CreateGroupResult

// DestroyGroupResult 删除分组结果
type DestroyGroupResult Result

// GetUserResult 获取用户详情
type GetUserResult struct {
	Result
	User model.UserDetailView `json:"user"`
}

// GetUserListResult 获取用户列表
type GetUserListResult struct {
	Result
	User []model.UserDetailView `json:"user"`
}

// CreateUserParam 新建用户参数
type CreateUserParam struct {
	Account  string        `json:"account"`
	Password string        `json:"password"`
	EMail    string        `json:"email"`
	Group    []model.Group `json:"group"`
}

// CreateUserResult 新建用户结果
type CreateUserResult struct {
	Result
	User model.UserDetailView `json:"user"`
}

// UpdateUserParam 更新用户参数
type UpdateUserParam struct {
	Email string        `json:"email"`
	Group []model.Group `json:"group"`
}

// UpdateUserPasswordParam 更新用户密码参数
type UpdateUserPasswordParam struct {
	Password string `json:"password"`
}

// UpdateUserResult 更新用户结果
type UpdateUserResult CreateUserResult

// DestroyUserResult 删除用户结果
type DestroyUserResult Result
