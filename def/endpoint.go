package def

import "github.com/muidea/magicCommon/model"

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
