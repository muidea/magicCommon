package def

import "muidea.com/magicCommon/model"

// GetModuleListResult 获取ModuleList
type GetModuleListResult struct {
	Result
	Module []model.ModuleDetailView `json:"module"`
}
