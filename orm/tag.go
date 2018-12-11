package orm

import (
	"fmt"
	"log"
	"strings"
)

type fieldTag interface {
	Name() string
	IsPrimaryKey() bool
	IsAutoIncrement() bool
	String() string
}

type tag struct {
	tagName         string
	isPrimaryKey    bool
	isAutoIncrement bool
}

// name[key][auto]
func newFieldTag(val string) fieldTag {
	items := strings.Split(val, " ")
	if len(items) < 1 {
		log.Fatalf("illegal tag value, value:%s", val)
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

	return &tag{tagName: tagName, isPrimaryKey: isPrimaryKey, isAutoIncrement: isAutoIncrement}
}

func (s *tag) Name() string {
	return s.tagName
}

func (s *tag) IsPrimaryKey() bool {
	return s.isPrimaryKey
}

func (s *tag) IsAutoIncrement() bool {
	return s.isAutoIncrement
}

func (s *tag) String() string {
	return fmt.Sprintf("%s key=%v auto=%v", s.tagName, s.isPrimaryKey, s.isAutoIncrement)
}
