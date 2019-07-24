package util

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	defaultPageSize = 10
	defaultPageNum  = 1
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
	if s.PageNum <= 0 {
		s.PageNum = 0
	}

	return true
}

// Encode compile
func (s *PageFilter) Encode() string {
	return fmt.Sprintf("pageSize=%d&pageNum=%d", s.PageSize, s.PageNum)
}
