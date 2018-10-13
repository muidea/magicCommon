package def

import "net/http"

// ContentFilter contentFilter
type ContentFilter struct {
	FilterValue string
}

// Parse 解析内容过滤值
func (s *ContentFilter) Parse(request *http.Request) bool {
	s.FilterValue = request.URL.Query().Get("contentFilter")
	return s.FilterValue != ""
}
