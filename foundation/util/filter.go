package util

import (
	"fmt"
	"net/http"
	"net/url"
)

// Filter value filter
type Filter interface {
	Filter(val interface{}) bool
}

// ContentFilter 过滤器
type ContentFilter struct {
	Pagination *Pagination
	ParamItems *ParamItems
}

// NewFilter new filter
func NewFilter() *ContentFilter {
	return &ContentFilter{Pagination: nil, ParamItems: &ParamItems{Items: map[string]string{}}}
}

// Decode 内容过滤器
func (s *ContentFilter) Decode(request *http.Request) bool {
	pageFilter := &Pagination{}
	if pageFilter.Decode(request) {
		s.Pagination = pageFilter
	}

	contentFilter := &ParamItems{}
	if contentFilter.Decode(request) {
		s.ParamItems = contentFilter
	}

	return s.Pagination != nil || s.ParamItems != nil
}

// Encode encode filter
func (s *ContentFilter) Encode(vals url.Values) url.Values {
	if s.Pagination != nil {
		vals = s.Pagination.Encode(vals)
	}
	if s.ParamItems != nil {
		vals = s.ParamItems.Encode(vals)
	}

	return vals
}

func (s *ContentFilter) Get(key string) (val string, ok bool) {
	if s.ParamItems != nil {
		val, ok = s.ParamItems.Items[key]
		return
	}

	return
}

func (s *ContentFilter) Set(key, value string) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = value
	}
}

func (s *ContentFilter) Remove(key string) {
	if s.ParamItems != nil {
		delete(s.ParamItems.Items, key)
	}
}

// ParamItems contentFilter
type ParamItems struct {
	Items map[string]string
}

// Decode 解析内容过滤值
func (s *ParamItems) Decode(request *http.Request) bool {
	s.Items = map[string]string{}
	vals := request.URL.Query()
	for k, v := range vals {
		s.Items[k] = v[0]
	}

	return true
}

// Encode ParamItems
func (s *ParamItems) Encode(vals url.Values) url.Values {
	for k, v := range s.Items {
		if v != "" {
			vals.Set(k, fmt.Sprintf("%s", v))
		}
	}

	return vals
}
