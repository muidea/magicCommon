package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultPageSize = 20
	defaultPageNum  = 1
)

// Pagination 页面过滤器
type Pagination struct {
	// 单页条目数
	PageSize int `json:"pageSize"`
	// 页码
	PageNum int `json:"pageNum"`
}

func NewPagination(defaultSize, defaultNum int) *Pagination {
	return &Pagination{PageSize: defaultSize, PageNum: defaultNum}
}

// Decode 从request里解析PageFilter
func (s *Pagination) Decode(request *http.Request) bool {
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
	s.PageNum = numValue
	if s.PageNum <= 0 {
		s.PageNum = 1
	}

	return true
}

// Encode encode url.Values
func (s *Pagination) Encode(vals url.Values) url.Values {
	vals.Set("pageSize", fmt.Sprintf("%d", s.PageSize))
	vals.Set("pageNum", fmt.Sprintf("%d", s.PageNum))

	return vals
}
