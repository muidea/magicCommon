package util

import (
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestContentFilter_Encode(t *testing.T) {
	filter := NewFilter()
	filter.Set("set", "123")
	filter.Equal("equal", "123")
	filter.NotEqual("notequal", "123")
	filter.Below("below", 40)
	filter.Above("above", 40)
	filter.In("in", []float32{12.23, 23.45})
	filter.NotIn("notin", []float32{12.23, 23.45})
	filter.Like("like", "hello world")

	val := url.Values{}
	val = filter.Encode(val)

	log.Print(val.Encode())

	req := &http.Request{URL: &url.URL{RawQuery: val.Encode()}}

	newFilter := NewFilter()
	newFilter.Decode(req)

	if !newFilter.ParamItems.IsEqual("set") {
		t.Error("invalid equal key")
		return
	}

	if !newFilter.ParamItems.IsEqual("equal") {
		t.Error("invalid equal key")
		return
	}
	if !newFilter.ParamItems.IsNotEqual("notequal") {
		t.Error("invalid not equal key")
		return
	}
	if !newFilter.ParamItems.IsBelow("below") {
		t.Error("invalid below key")
		return
	}
	if !newFilter.ParamItems.IsAbove("above") {
		t.Error("invalid above key")
		return
	}
	if !newFilter.ParamItems.IsIn("in") {
		t.Error("invalid in key")
		return
	}
	if !newFilter.ParamItems.IsNotIn("notin") {
		t.Error("invalid not in key")
		return
	}
	if !newFilter.ParamItems.IsLike("like") {
		t.Error("invalid like key")
		return
	}
}
