package model

import (
	"reflect"
	"testing"
)

func TestValue(t *testing.T) {
	type AA struct {
		ii int
		jj int
	}
	val := []AA{AA{ii: 12, jj: 23}}
	fv := newFieldValue(reflect.ValueOf(val))

	fv.GetDepend()
}
