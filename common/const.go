package common

import "muidea.com/magicCommon/model"

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

// VisitorAuthGroup 访客权限组
var VisitorAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 0, Name: "访客权限组"}, Description: "允许查看公开权限的内容，无须登录"}

// UserAuthGroup 用户权限组
var UserAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 1, Name: "用户权限组"}, Description: "允许查看用户权限的内容以及公开权限的内容，要求预先进行登录"}

// MaintainerAuthGroup 维护权限组
var MaintainerAuthGroup = model.AuthGroup{Unit: model.Unit{ID: 2, Name: "维护权限组"}, Description: "允许查看和编辑内容，要求预先进行登录"}

// UnknownAuthGroup 未知授权组
var UnknownAuthGroup = model.AuthGroup{Unit: model.Unit{ID: -1, Name: "未知权限组"}, Description: "不是合法的授权组，原因是由于查询提供的ID不是有效值"}

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

// DefaultContentCatalog 系统默认的Content分组，UpdataCatalog时，如果需要创建Catalog,则默认指定的ParentCatalog
var DefaultContentCatalog = model.CatalogDetail{Summary: model.Summary{Unit: model.Unit{ID: 0, Name: "默认Content分组"}, CreateDate: "", Creater: 0}, Description: "系统默认的Content分组"}

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
