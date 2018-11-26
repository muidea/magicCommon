package orm

import (
	"fmt"
)

// single field info
type fieldInfo struct {
	mi         *modelInfo
	fieldIndex int
	fieldName  string
	fieldType  int
}

// field info collection
type fields struct {
	pk     *fieldInfo
	fields map[string]*fieldInfo
}

func (s *fieldInfo) Dump() string {
	return fmt.Sprintf("index:%d,name:%s,type:%d\n", s.fieldIndex, s.fieldName, s.fieldType)
}

func (s *fields) Dump() {
	if s.pk != nil {
		fmt.Printf("pk:[%s]\n", s.pk.Dump())
	}
	for k, v := range s.fields {
		fmt.Printf("name:%s,item:[%s]\n", k, v.Dump())
	}
}
