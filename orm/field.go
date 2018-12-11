package orm

import (
	"fmt"
	"log"
	"reflect"
)

// single field info
type fieldInfo struct {
	fieldIndex int
	fieldName  string
	fieldType  int
	fieldTag   fieldTag
}

// field info collection
type fields struct {
	pk *fieldInfo

	// name->fieldInfo
	fields map[string]*fieldInfo
}

func (s *fieldInfo) isReference() bool {
	return false
}

func (s *fieldInfo) Dump() string {
	return fmt.Sprintf("index:%d,name:%s,type:%d,tag:%s", s.fieldIndex, s.fieldName, s.fieldType, s.fieldTag)
}

func (s *fields) append(sf *fieldInfo) {
	_, ok := s.fields[sf.fieldName]
	if ok {
		log.Fatalf("duplicate field,[%s]", sf.Dump())
	}

	s.fields[sf.fieldName] = sf
}

func (s *fields) Dump() {
	if s.pk != nil {
		fmt.Printf("pk:[%s]\n", s.pk.Dump())
	}

	for k, v := range s.fields {
		fmt.Printf("key:%s, val:[%s]\n", k, v.Dump())
	}
}

func getFieldInfo(idx int, sf *reflect.StructField, sv *reflect.Value) *fieldInfo {
	info := &fieldInfo{}
	info.fieldIndex = idx
	info.fieldName = sf.Name
	info.fieldTag = newFieldTag(sf.Tag.Get("orm"))
	tVal, err := getFieldType(*sv)
	if err != nil {
		log.Printf("getFieldType failed, idx:%d, name:%s, type:%s, err:%s", idx, sf.Name, sf.Type.Name(), err.Error())
		return nil
	}
	log.Printf("idx:%d, name:%s, type:%s", idx, sf.Name, sf.Type.Name())
	info.fieldType = tVal

	return info
}
