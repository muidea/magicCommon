package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func DefaultPagination() *Pagination {
	return &Pagination{PageSize: defaultPageSize, PageNum: defaultPageNum}
}

func (s *Pagination) String() string {
	ss := strings.Builder{}
	ss.WriteString(fmt.Sprintf("[%d:%d]", s.PageSize, s.PageNum))
	return ss.String()
}

// Decode 从request里解析PageFilter
func (s *Pagination) Decode(request *http.Request) bool {
	pageSize := request.URL.Query().Get("_pageSize")
	pageNum := request.URL.Query().Get("_pageNum")
	if pageSize == "" && pageNum == "" {
		return false
	}

	sizeValue, err := strconv.Atoi(pageSize)
	if err != nil || sizeValue == 0 {
		sizeValue = defaultPageSize
	}
	s.PageSize = sizeValue

	numValue, err := strconv.Atoi(pageNum)
	if err != nil || numValue <= 0 {
		numValue = defaultPageNum
	}
	s.PageNum = numValue

	return true
}

// Encode encode url.Values
func (s *Pagination) Encode(vals url.Values) url.Values {
	vals.Set("_pageSize", fmt.Sprintf("%d", s.PageSize))
	vals.Set("_pageNum", fmt.Sprintf("%d", s.PageNum))

	return vals
}
