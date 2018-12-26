package model

import (
	"fmt"
	"reflect"

	"muidea.com/magicCommon/orm/util"
)

// FieldType FieldType
type FieldType interface {
	Name() string
	Value() int
	Catalog() int
	IsPtr() bool
	PkgPath() string
	String() string
}

type typeImpl struct {
	typeValue   int
	typeName    string
	typePkgPath string
	typeCatalog int
	typeIsPtr   bool
}

func (s *typeImpl) Name() string {
	return s.typeName
}

func (s *typeImpl) Value() int {
	return s.typeValue
}

func (s *typeImpl) Catalog() int {
	return s.typeCatalog
}

func (s *typeImpl) IsPtr() bool {
	return s.typeIsPtr
}

func (s *typeImpl) PkgPath() string {
	return s.typePkgPath
}

func (s *typeImpl) String() string {
	return fmt.Sprintf("val:%d,name:%s,pkgPath:%s, catalog:%v", s.typeValue, s.typeName, s.typePkgPath, s.typeCatalog)
}

func newFieldType(sf *reflect.StructField) (ret FieldType, err error) {
	val := sf.Type

	isPtr := false
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		isPtr = true
	}

	tVal, tErr := GetValueTypeEnum(val)
	if tErr != nil {
		err = tErr
		return
	}

	tCatalog := util.TypeBaseTypeField
	isReference := IsReferenceType(val)
	if isReference {
		tCatalog = util.TypeReferenceField
		if isPtr {
			tCatalog = util.TypeReferencePtrField
		}
	}

	ret = &typeImpl{typeValue: tVal, typeName: val.String(), typePkgPath: val.PkgPath(), typeCatalog: tCatalog, typeIsPtr: isPtr}
	return
}
