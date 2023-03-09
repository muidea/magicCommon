package util

// SortFilter 页面过滤器
type SortFilter struct {
	// true:升序,false:降序
	AscFlag bool `json:"ascFlag"`
	// 排序字段
	FieldName string `json:"fieldName"`
}

func NewSortFilter(name string, ascFlag bool) *SortFilter {
	return &SortFilter{AscFlag: ascFlag, FieldName: name}
}

// Name return name
func (s *SortFilter) Name() string {
	return s.FieldName
}

func (s *SortFilter) AscSort() bool {
	return s.AscFlag
}
