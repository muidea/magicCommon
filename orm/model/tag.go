package model

import (
	"fmt"
	"strings"
)

// FieldTag FieldTag
type FieldTag interface {
	Name() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	String() string
	Copy() FieldTag
}

type tagImpl struct {
	tagName         string
	isPrimaryKey    bool
	isAutoIncrement bool
}

// name[key][auto]
func newFieldTag(val string) (ret FieldTag, err error) {
	items := strings.Split(val, " ")
	if len(items) < 1 {
		err = fmt.Errorf("illegal tagImpl value, value:%s", val)
		return
	}

	tagName := items[0]
	isPrimaryKey := false
	isAutoIncrement := false
	if len(items) >= 2 {
		switch items[1] {
		case "key":
			isPrimaryKey = true
		case "auto":
			isAutoIncrement = true
		}
	}
	if len(items) >= 3 {
		switch items[2] {
		case "key":
			isPrimaryKey = true
		case "auto":
			isAutoIncrement = true
		}
	}

	ret = &tagImpl{tagName: tagName, isPrimaryKey: isPrimaryKey, isAutoIncrement: isAutoIncrement}
	return
}

func (s *tagImpl) Name() string {
	return s.tagName
}

func (s *tagImpl) IsPrimaryKey() bool {
	return s.isPrimaryKey
}

func (s *tagImpl) IsAutoIncrement() bool {
	return s.isAutoIncrement
}

func (s *tagImpl) String() string {
	return fmt.Sprintf("name=%s key=%v auto=%v", s.tagName, s.isPrimaryKey, s.isAutoIncrement)
}

func (s *tagImpl) Copy() FieldTag {
	return &tagImpl{
		tagName:         s.tagName,
		isPrimaryKey:    s.isPrimaryKey,
		isAutoIncrement: s.isAutoIncrement,
	}
}
