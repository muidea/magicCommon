package def

import "muidea.com/magicCommon/model"

// QuerySystemConfigResult 查询SystemConfig结果
type QuerySystemConfigResult struct {
	Result
	SystemProperty model.SystemProperty `json:"systemProperty"`
}

// UpdateSystemConfigResult 更新SystemConfig结果
type UpdateSystemConfigResult Result
