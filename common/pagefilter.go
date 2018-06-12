package common

import (
	"net/http"
	"strconv"
)

const (
	defaultPageSize   = 10
	defaultPagination = 1
)

// PageFilter 页面过滤器
type PageFilter struct {
	// 单页条目数
	PageSize int `json:"pageSize"`
	// 页码
	Pagination int `json:"pagination"`
}

// Parse 从request里解析PageFilter
func (s *PageFilter) Parse(request *http.Request) {
	pageSize := request.URL.Query().Get("pageSize")
	sizeValue, err := strconv.Atoi(pageSize)
	if err != nil {
		sizeValue = defaultPageSize
	}
	s.PageSize = sizeValue

	pagination := request.URL.Query().Get("pagination")
	paginationValue, err := strconv.Atoi(pagination)
	if err != nil {
		paginationValue = defaultPagination
	}
	s.Pagination = paginationValue
}
