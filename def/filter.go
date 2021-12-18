package def

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/muidea/magicCommon/foundation/util"
)

// Filter 过滤器
type Filter struct {
	PageFilter    *util.PageFilter
	ContentFilter *ContentFilter
}

// NewFilter new filter
func NewFilter() *Filter {
	return &Filter{PageFilter: nil, ContentFilter: &ContentFilter{Items: map[string]string{}}}
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
func (s *Filter) Encode(vals url.Values) url.Values {
	if s.PageFilter != nil {
		vals = s.PageFilter.Encode(vals)
	}
	if s.ContentFilter != nil {
		vals = s.ContentFilter.Encode(vals)
	}

	return vals
}

func (s *Filter) Get(key string) (val string, ok bool) {
	if s.ContentFilter != nil {
		val, ok = s.ContentFilter.Items[key]
		return
	}

	return
}

func (s *Filter) Set(key, value string) {
	if s.ContentFilter != nil {
		s.ContentFilter.Items[key] = value
	}
}

func (s *Filter) Remove(key string) {
	if s.ContentFilter != nil {
		delete(s.ContentFilter.Items, key)
	}
}

// ContentFilter contentFilter
type ContentFilter struct {
	Items map[string]string
}

// Decode 解析内容过滤值
func (s *ContentFilter) Decode(request *http.Request) bool {
	s.Items = map[string]string{}
	vals := request.URL.Query()
	for k, v := range vals {
		s.Items[k] = v[0]
	}

	return true
}

// Encode ContentFilter
func (s *ContentFilter) Encode(vals url.Values) url.Values {
	for k, v := range s.Items {
		if v != "" {
			vals.Set(k, fmt.Sprintf("%s", v))
		}
	}

	return vals
}
