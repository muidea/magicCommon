package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Filter value filter
type Filter interface {
	Filter(val interface{}) bool
}

// ContentFilter 过滤器
type ContentFilter struct {
	PaginationPtr *Pagination `json:"pagination"`
	ParamItems    *ParamItems `json:"params"`
}

// NewFilter new filter
func NewFilter(name, pkgPath string) *ContentFilter {
	return &ContentFilter{PaginationPtr: nil, ParamItems: &ParamItems{Name: name, PkgPath: pkgPath, Items: map[string]string{}}}
}

// Decode 内容过滤器
func (s *ContentFilter) Decode(request *http.Request) {
	pagePtr := &Pagination{}
	if pagePtr.Decode(request) {
		s.PaginationPtr = pagePtr
	}

	paramPtr := &ParamItems{}
	if paramPtr.Decode(request) {
		s.ParamItems = paramPtr
	}
}

// Encode encode filter
func (s *ContentFilter) Encode(vals url.Values) url.Values {
	if s.PaginationPtr != nil {
		vals = s.PaginationPtr.Encode(vals)
	}
	if s.ParamItems != nil {
		vals = s.ParamItems.Encode(vals)
	}

	return vals
}

func (s *ContentFilter) GetName() string {
	if s.ParamItems != nil {
		return s.ParamItems.Name
	}

	return ""
}

func (s *ContentFilter) GetPkgPath() string {
	if s.ParamItems != nil {
		return s.ParamItems.PkgPath
	}

	return ""
}

func (s *ContentFilter) GetPkeKey() string {
	if s.ParamItems != nil {
		return path.Join(s.ParamItems.PkgPath, s.ParamItems.Name)
	}

	return ""
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
		s.ParamItems.Equal(key, value)
	}
}

func (s *ContentFilter) GetEqual(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetEqual(key)
	}

	return nil
}

func (s *ContentFilter) NotEqual(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.NotEqual(key, value)
	}
}

func (s *ContentFilter) GetNotEqual(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetNotEqual(key)
	}

	return nil
}

func (s *ContentFilter) Below(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Below(key, value)
	}
}

func (s *ContentFilter) GetBelow(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetBelow(key)
	}

	return nil
}

func (s *ContentFilter) Above(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Above(key, value)
	}
}

func (s *ContentFilter) GetAbove(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetAbove(key)
	}

	return nil
}

func (s *ContentFilter) In(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.In(key, value)
	}
}

func (s *ContentFilter) GetIn(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetIn(key)
	}

	return nil
}

func (s *ContentFilter) NotIn(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.NotIn(key, value)
	}
}

func (s *ContentFilter) GetNotIn(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetNotIn(key)
	}

	return nil
}

func (s *ContentFilter) Like(key string, value interface{}) {
	if s.ParamItems != nil {
		s.ParamItems.Like(key, value)
	}
}

func (s *ContentFilter) GetLike(key string) interface{} {
	if s.ParamItems != nil {
		return s.ParamItems.GetLike(key)
	}

	return nil
}

func (s *ContentFilter) Remove(key string) {
	if s.ParamItems != nil {
		delete(s.ParamItems.Items, key)
	}
}

func (s *ContentFilter) Pagination(ptr *Pagination) {
	s.PaginationPtr = ptr
}

func (s *ContentFilter) SortFilter(ptr *SortFilter) {
	if s.ParamItems != nil {
		s.ParamItems.SortFilter = ptr
	}
}

func (s *ContentFilter) BindEntity(name, pkgPath string) {
	if s.ParamItems != nil {
		s.ParamItems.Name = name
		s.ParamItems.PkgPath = pkgPath
	}
}

func (s *ContentFilter) ValueMask(val any) {
	if s.ParamItems != nil {
		byteVal, byteErr := json.Marshal(val)
		if byteErr != nil {
			return
		}

		s.ParamItems.ValueMask = byteVal
	}
}

// ParamItems contentFilter
type ParamItems struct {
	Name       string            `json:"name"`
	PkgPath    string            `json:"pkgPath"`
	Items      map[string]string `json:"items"`
	SortFilter *SortFilter       `json:"sortFilter"`
	ValueMask  json.RawMessage   `json:"valueMask"`
}

const sortKey = "_sort"

// Decode 解析内容过滤值
func (s *ParamItems) Decode(request *http.Request) bool {
	s.Items = map[string]string{}
	vals := request.URL.Query()
	for k, v := range vals {
		if strings.HasPrefix(k, "_") {
			if k == sortKey {
				ptr := &SortFilter{}
				err := json.Unmarshal([]byte(vals.Get(sortKey)), ptr)
				if err == nil {
					s.SortFilter = ptr
				}
			}

			continue
		}

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

	if s.SortFilter != nil {
		byteVal, byteErr := json.Marshal(s.SortFilter)
		if byteErr == nil {
			vals.Set(sortKey, string(byteVal))
		}
	}

	return vals
}

func (s *ParamItems) Equal(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|=", MarshalString(value))
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

	return GetEqual(val)
}

func (s *ParamItems) NotEqual(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|!=", MarshalString(value))
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

	return GetNotEqual(val)
}

func (s *ParamItems) Below(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|<", MarshalString(value))
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

	return GetBelow(val)
}

func (s *ParamItems) Above(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|>", MarshalString(value))
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

	return GetAbove(val)
}

func (s *ParamItems) In(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|in", MarshalString(value))
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

	return GetIn(val)
}

func (s *ParamItems) NotIn(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|!in", MarshalString(value))
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

	return GetNotIn(val)
}

func (s *ParamItems) Like(key string, value interface{}) {
	s.Items[key] = fmt.Sprintf("%v|like", value)
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

	return GetLike(val)
}

func GetEqual(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return UnmarshalString(val)
	}
	if val[idx+1:] != "=" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetNotEqual(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "!=" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetBelow(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "<" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetAbove(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != ">" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetIn(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "in" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetNotIn(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "!in" {
		return nil
	}

	return UnmarshalString(val[:idx])
}

func GetLike(val string) interface{} {
	idx := strings.LastIndex(val, "|")
	if idx == -1 {
		return nil
	}
	if val[idx+1:] != "like" {
		return nil
	}

	return val[:idx]
}
