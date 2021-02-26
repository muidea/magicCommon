package util

import "fmt"

// SortFilter 页面过滤器
type SortFilter struct {
	// true:升序,false:降序
	AscSort bool `json:"ascSort"`
	// 排序字段
	FieldName string `json:"fieldName"`
}

func NewSortFilter(name string, ascSort bool) *SortFilter {
	return &SortFilter{AscSort: ascSort, FieldName: name}
}

// Name return name
func (s *SortFilter) Name() string {
	return s.FieldName
}

// SortStr return sort string
func (s *SortFilter) SortStr(tagName string) string {
	if tagName == "" {
		tagName = s.FieldName
	}

	if s.AscSort {
		return fmt.Sprintf("%s asc", tagName)
	}

	return fmt.Sprintf("%s desc", tagName)
}
