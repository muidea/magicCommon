package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultPageSize = 10
	defaultPageNum  = 1
)

// Pagination 页面过滤器
type Pagination struct {
	// 单页条目数
	PageSize int `json:"page_size"`
	// 页码
	PageNum int `json:"page_num"`
}

func NewPagination() *Pagination {
	return &Pagination{}
}

// Decode 从request里解析PageFilter
func (s *Pagination) Decode(request *http.Request) bool {
	pageSize := request.URL.Query().Get("page_size")
	pageNum := request.URL.Query().Get("page_num")
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
	vals.Set("page_size", fmt.Sprintf("%d", s.PageSize))
	vals.Set("page_num", fmt.Sprintf("%d", s.PageNum))

	return vals
}
