package util

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
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
		val = strings.Trim(val, "\"")
		return
	}

	return
}

func (s *ContentFilter) Set(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = MarshalString(value)
	}
}

func (s *ContentFilter) Equal(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|=", MarshalString(value))
	}
}

func (s *ContentFilter) NotEqual(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|!=", MarshalString(value))
	}
}

func (s *ContentFilter) Below(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|<", MarshalString(value))
	}
}

func (s *ContentFilter) Above(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|>", MarshalString(value))
	}
}

func (s *ContentFilter) In(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|in", MarshalString(value))
	}
}

func (s *ContentFilter) NotIn(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|!in", MarshalString(value))
	}
}

func (s *ContentFilter) Like(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Items[key] = fmt.Sprintf("%v|like", value)
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

func (s *ParamItems) IsEqual(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return true
	}
	return val[idx+1:] == "="
}

func (s *ParamItems) GetEqual(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return UnmarshalString(val)
	}
	if val[idx+1:] != "=" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsNotEqual(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return false
	}
	return val[idx+1:] == "!="
}

func (s *ParamItems) GetNotEqual(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "!=" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsBelow(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return true
	}
	return val[idx+1:] == "<"
}

func (s *ParamItems) GetBelow(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "<" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsAbove(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return false
	}
	return val[idx+1:] == ">"
}

func (s *ParamItems) GetAbove(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != ">" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsIn(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return true
	}
	return val[idx+1:] == "in"
}

func (s *ParamItems) GetIn(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "in" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsNotIn(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return false
	}
	return val[idx+1:] == "!in"
}

func (s *ParamItems) GetNotIn(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "!in" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func (s *ParamItems) IsLike(key string) bool {
	val, ok := s.Items[key]
	if !ok {
		return false
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return false
	}
	return val[idx+1:] == "like"
}

func (s *ParamItems) GetLike(key string) interface{} {
	val, ok := s.Items[key]
	if !ok {
		return nil
	}

	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "like" {
		return nil
	}

	return val[:idx]
}
