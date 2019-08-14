package def

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/muidea/magicCommon/foundation/util"
)

// Filter 过滤器
type Filter struct {
	items         []string
	PageFilter    *util.PageFilter
	ContentFilter *ContentFilter
}

// NewFilter new filter
func NewFilter(items []string) *Filter {
	return &Filter{items: items}
}

// Decode 内容过滤器
func (s *Filter) Decode(request *http.Request) bool {
	pageFilter := &util.PageFilter{}
	if pageFilter.Decode(request) {
		s.PageFilter = pageFilter
	}

	contentFilter := &ContentFilter{}
	if contentFilter.Decode(request, s.items) {
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
	Items map[string]string
}

// Decode 解析内容过滤值
func (s *ContentFilter) Decode(request *http.Request, items []string) bool {
	s.Items = map[string]string{}
	for _, k := range items {
		val := request.URL.Query().Get(k)
		if val != "" {
			s.Items[k] = val
		}
	}

	return true
}

// Encode ContentFilter
func (s *ContentFilter) Encode() string {
	val := url.Values{}
	for k, v := range s.Items {
		if v != "" {
			val.Set(k, v)
		}
	}

	return val.Encode()
}
