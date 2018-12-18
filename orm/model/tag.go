package model

import (
	"fmt"
	"log"
	"strings"
)

// FieldTag FieldTag
type FieldTag interface {
	Name() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	String() string
}

type tagImpl struct {
	tagName         string
	isPrimaryKey    bool
	isAutoIncrement bool
}

// name[key][auto]
func newFieldTag(val string) FieldTag {
	items := strings.Split(val, " ")
	if len(items) < 1 {
		log.Fatalf("illegal tagImpl value, value:%s", val)
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

	return &tagImpl{tagName: tagName, isPrimaryKey: isPrimaryKey, isAutoIncrement: isAutoIncrement}
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
	return fmt.Sprintf("%s key=%v auto=%v", s.tagName, s.isPrimaryKey, s.isAutoIncrement)
}
