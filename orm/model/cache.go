package model

import "muidea.com/magicCommon/foundation/cache"

// StructInfoCache StructInfo Cache
type StructInfoCache interface {
	Reset()

	Put(info *StructInfo)

	Find(name string) *StructInfo
}

type impl struct {
	kvCache cache.KVCache
}

// NewCache new structInfo cache
func NewCache() StructInfoCache {
	return &impl{kvCache: cache.NewKVCache()}
}

func (s *impl) Reset() {
	s.kvCache.ClearAll()
}

func (s *impl) Put(info *StructInfo) {
	s.kvCache.Put(info.GetStructName(), info, cache.MaxAgeValue)
}

func (s *impl) Find(name string) *StructInfo {
	obj, ok := s.kvCache.Fetch(name)
	if !ok {
		return nil
	}

	return obj.(*StructInfo)
}
