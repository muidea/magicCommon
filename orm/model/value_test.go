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
	val := []*AA{&AA{ii: 12, jj: 23}, &AA{ii: 23, jj: 34}}
	fv := newFieldValue(reflect.ValueOf(val))

	fds := fv.GetDepend()
	if len(fds) != 2 {
		t.Errorf("fv.GetDepend failed. fds size:%d", len(fds))
	}
}
