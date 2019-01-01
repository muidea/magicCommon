package model

import "muidea.com/magicCommon/foundation/cache"

// StructInfoCache StructInfo Cache
type StructInfoCache interface {
	Reset()

	Put(name string, info StructInfo)

	Fetch(name string) StructInfo

	Remove(name string)
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

func (s *impl) Put(name string, info StructInfo) {
	s.kvCache.Put(name, info, cache.MaxAgeValue)
}

func (s *impl) Fetch(name string) StructInfo {
	obj, ok := s.kvCache.Fetch(name)
	if !ok {
		return nil
	}

	return obj.(StructInfo)
}

func (s *impl) Remove(name string) {
	s.kvCache.Remove(name)
}
