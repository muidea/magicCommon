package common

import "muidea.com/magicCommon/model"

// SessionID 会话ID
const SessionID = "sessionID"

// AuthTokenID 鉴权Token
const AuthTokenID = "authToken"

// ExpiryDate 会话有效期
const ExpiryDate = "expiryDate"

const (
	// NEW 新建状态
	NEW = iota
	// ACTIVE 激活
	ACTIVE
	// DEACTIVE 未激活
	DEACTIVE
	// DISABLE 禁用
	DISABLE
)

// BuildinAccountGroup 系统内置分组
var BuildinAccountGroup = model.GroupDetail{Group: model.Group{ID: 0, Name: "系统基础分组"}, Description: "系统基础分组", Catalog: 0}

// BuildinAccountUser 系统内置用户
var BuildinAccountUser = model.UserDetail{User: model.User{ID: 0, Name: "admin"}, Email: "rangh@foxmail.com", Group: []int{0}, Status: ACTIVE}

// VisitorAuthGroup 访客权限组
var VisitorAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 0, Name: "访客组权限"}, Description: "允许查看公开权限的内容，无须登录"}

// UserAuthGroup 用户权限组
var UserAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 1, Name: "用户组权限"}, Description: "允许查看用户权限的内容以及公开权限的内容，要求预先进行登录"}

// MaintainerAuthGroup 维护权限组
var MaintainerAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 2, Name: "维护组权限"}, Description: "允许查看和编辑内容，要求预先进行登录"}

// UnknownAuthGroup 未知授权组
var UnknownAuthGroup = model.AuthGroup{Unit: model.Unit{ID: -1, Name: "未知权限"}, Description: "不是合法的授权组，原因是由于查询提供的ID不是有效值"}

// GetAuthGroup 获取指定授权组
func GetAuthGroup(id int) model.AuthGroup {
	switch id {
	case VisitorAuthGroup.ID:
		return VisitorAuthGroup
	case UserAuthGroup.ID:
		return UserAuthGroup
	case MaintainerAuthGroup.ID:
		return MaintainerAuthGroup
	default:
		return UnknownAuthGroup
	}
}

// BuildinContentCatalog 系统默认的Content分组，UpdataCatalog时，如果需要创建Catalog,则默认指定的ParentCatalog
var BuildinContentCatalog = model.CatalogDetail{Summary: model.Summary{Unit: model.Unit{ID: 0, Name: "Content基础分类"}, Description: "Content基础分类", Type: model.CATALOG}}

// NewStatus 新建状态
var NewStatus = model.Status{ID: NEW, Name: "新建"}

// ActiveStatus 激活状态
var ActiveStatus = model.Status{ID: ACTIVE, Name: "激活"}

// DeactiveStatus 未激活状态
var DeactiveStatus = model.Status{ID: DEACTIVE, Name: "未激活"}

// DisableStatus 未激活状态
var DisableStatus = model.Status{ID: DISABLE, Name: "禁用"}

// UnknownStatus 未知状态
var UnknownStatus = model.Status{Name: "未知"}

// GetStatus 获取指定状态
func GetStatus(id int) model.Status {
	switch id {
	case NEW:
		return NewStatus
	case ACTIVE:
		return ActiveStatus
	case DEACTIVE:
		return DeactiveStatus
	case DISABLE:
		return DisableStatus
	default:
		status := UnknownStatus
		status.ID = id
		return status
	}
}

// 模块类型
const (
	// 内核模块，不能被禁用
	KERNEL = iota
	// 内置模块，属于系统自带可选模块，可以被禁用
	INTERNAL
	// 外部模块，通过外部接口注册进来的模块，可以被禁用
	EXTERNAL
)

// KernelModule 内核模块
var KernelModule = model.ModuleType{ID: KERNEL, Name: "内核模块"}

// InternalModule 内置模块
var InternalModule = model.ModuleType{ID: INTERNAL, Name: "内置模块"}

// ExternalModule 外部模块
var ExternalModule = model.ModuleType{ID: EXTERNAL, Name: "外部模块"}

// InvalidModule 非法模块
var InvalidModule = model.ModuleType{Name: "非法模块"}

// GetModuleType 获取模块类型
func GetModuleType(id int) model.ModuleType {
	switch id {
	case KERNEL:
		return KernelModule
	case INTERNAL:
		return InternalModule
	case EXTERNAL:
		return ExternalModule
	default:
		moduleType := InvalidModule
		moduleType.ID = id
		return moduleType
	}
}
