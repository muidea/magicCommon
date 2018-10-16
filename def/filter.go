package def

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Filter 过滤器
type Filter struct {
	PageFilter    *PageFilter
	ContentFilter *ContentFilter
}

// Decode 内容过滤器
func (s *Filter) Decode(request *http.Request) bool {
	pageFilter := &PageFilter{}
	if pageFilter.Decode(request) {
		s.PageFilter = pageFilter
	}

	contentFilter := &ContentFilter{}
	if contentFilter.Decode(request) {
		s.ContentFilter = contentFilter
	}

	return s.PageFilter != nil || s.ContentFilter != nil
}

// ContentFilter contentFilter
type ContentFilter struct {
	FilterValue string
}

// Decode 解析内容过滤值
func (s *ContentFilter) Decode(request *http.Request) bool {
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

// Decode 从request里解析PageFilter
func (s *PageFilter) Decode(request *http.Request) bool {
	pageSize := request.URL.Query().Get("pageSize")
	pageNum := request.URL.Query().Get("pageNum")
	if pageSize == "" && pageNum == "" {
		return false
	}

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

// Encode compile
func (s *PageFilter) Encode() string {
	return url.QueryEscape(fmt.Sprintf("pageSize=%d&pageNum=%d", s.PageSize, s.PageNum))
}
