package def

import (
	"net/http"
	"strconv"
)

// Filter 过滤器
type Filter struct {
	PageFilter    *PageFilter
	ContentFilter *ContentFilter
}

// Parse 内容过滤器
func (s *Filter) Parse(request *http.Request) {
	pageFilter := &PageFilter{}
	if pageFilter.Parse(request) {
		s.PageFilter = pageFilter
	}

	contentFilter := &ContentFilter{}
	if contentFilter.Parse(request) {
		s.ContentFilter = contentFilter
	}
}

// ContentFilter contentFilter
type ContentFilter struct {
	FilterValue string
}

// Parse 解析内容过滤值
func (s *ContentFilter) Parse(request *http.Request) bool {
	s.FilterValue = request.URL.Query().Get("filterValue")
	return s.FilterValue != ""
}

const (
	defaultPageSize = 10
	defaultPageNum  = 0
)

// PageFilter 页面过滤器
type PageFilter struct {
	// 单页条目数
	PageSize int `json:"pageSize"`
	// 页码
	PageNum int `json:"pageNum"`
}

// Parse 从request里解析PageFilter
func (s *PageFilter) Parse(request *http.Request) bool {
	pageSize := request.URL.Query().Get("pageSize")
	pageNum := request.URL.Query().Get("pageNum")
	sizeValue, err := strconv.Atoi(pageSize)
	if err != nil {
		sizeValue = defaultPageSize
	}
	s.PageSize = sizeValue

	numValue, err := strconv.Atoi(pageNum)
	if err != nil {
		numValue = defaultPageNum
	}
	s.PageNum = numValue - 1
	if s.PageNum < 0 {
		s.PageNum = 0
	}

	return true
}
