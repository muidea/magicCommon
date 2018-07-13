package def

import (
	"net/http"
	"strconv"
)

const (
	defaultPageSize = 10
	defaultPage     = 1
)

// PageFilter 页面过滤器
type PageFilter struct {
	// 单页条目数
	PageSize int `json:"pageSize"`
	// 页码
	Page int `json:"page"`
}

// Parse 从request里解析PageFilter
func (s *PageFilter) Parse(request *http.Request) {
	pageSize := request.URL.Query().Get("pageSize")
	sizeValue, err := strconv.Atoi(pageSize)
	if err != nil {
		sizeValue = defaultPageSize
	}
	s.PageSize = sizeValue

	page := request.URL.Query().Get("page")
	pageValue, err := strconv.Atoi(page)
	if err != nil {
		pageValue = defaultPage
	}
	s.Page = pageValue
}
