package def

import (
	"fmt"
	"net/http"

	"muidea.com/magicCommon/foundation/util"
)

// Filter 过滤器
type Filter struct {
	PageFilter    *util.PageFilter
	ContentFilter *ContentFilter
}

// Decode 内容过滤器
func (s *Filter) Decode(request *http.Request) bool {
	pageFilter := &util.PageFilter{}
	if pageFilter.Decode(request) {
		s.PageFilter = pageFilter
	}

	contentFilter := &ContentFilter{}
	if contentFilter.Decode(request) {
		s.ContentFilter = contentFilter
	}

	return s.PageFilter != nil || s.ContentFilter != nil
}

// Encode encode filter
func (s *Filter) Encode() string {
	retVal := ""
	if s.PageFilter != nil {
		retVal = fmt.Sprintf("%s", s.PageFilter.Encode())
	}
	if s.ContentFilter != nil {
		if retVal == "" {
			retVal = fmt.Sprintf("%s", s.ContentFilter.Encode())
		} else {
			retVal = fmt.Sprintf("%s&%s", retVal, s.ContentFilter.Encode())
		}
	}

	return retVal
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

// Encode ContentFilter
func (s *ContentFilter) Encode() string {
	if s.FilterValue == "" {
		return ""
	}

	return fmt.Sprintf("filterValue=%s", s.FilterValue)
}
