package def

import "github.com/muidea/magicCommon/model"

// GetModuleListResult 获取ModuleList
type GetModuleListResult struct {
	Result
	Module []model.ModuleDetailView `json:"module"`
}
