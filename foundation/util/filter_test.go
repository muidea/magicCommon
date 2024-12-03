package util

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
)

func TestContentFilter_Encode(t *testing.T) {
	filter := NewFilter("name", "pkgPath")
	filter.Set("set", -123)
	filter.Set("setB", true)
	filter.Set("setHey", "hey")
	filter.Equal("equal", -123)
	filter.NotEqual("notequal", "123")
	filter.Below("below", 40)
	filter.Above("above", 40)
	filter.In("in", []float32{12.23, 23.45})
	filter.In("inB", []bool{true, false, true})
	filter.NotIn("notin", []float32{12.23, 23.45})
	filter.Like("like", "hello world")

	byteVal, byteErr := json.Marshal(filter)
	if byteErr != nil {
		t.Error("encode json value failed")
		return
	}

	nf := NewFilter("name", "pkgPath")
	byteErr = json.Unmarshal(byteVal, nf)
	if byteErr != nil {
		t.Error("decode json value failed")
		return
	}
	checkFilter(t, nf)

	val := url.Values{}
	val = filter.Encode(val)

	log.Infof(val.Encode())

	req := &http.Request{URL: &url.URL{RawQuery: val.Encode()}}

	newFilter := NewFilter("name", "pkgPath")
	newFilter.Decode(req)

	checkFilter(t, newFilter)
}

func checkFilter(t *testing.T, newFilter *ContentFilter) {
	gVal, gOk := newFilter.Get("set")
	if !gOk {
		t.Error("invalid get key")
		return
	}
	if gVal != "-123" {
		t.Error("invalid get key")
		return
	}

	gVal, gOk = newFilter.Get("setHey")
	if !gOk {
		t.Error("invalid get key")
		return
	}
	if gVal != "hey" {
		t.Error("invalid get key")
		return
	}

	if !newFilter.ParamItems.IsEqual("set") {
		t.Error("invalid equal key")
		return
	}
	if newFilter.ParamItems.GetEqual("set") == nil {
		t.Error("invalid equal key")
		return
	}

	if !newFilter.ParamItems.IsEqual("setB") {
		t.Error("invalid equal key")
		return
	}
	if newFilter.ParamItems.GetEqual("setB") == nil {
		t.Error("invalid equal key")
		return
	}

	if !newFilter.ParamItems.IsEqual("equal") {
		t.Error("invalid equal key")
		return
	}
	if newFilter.ParamItems.GetEqual("equal") == nil {
		t.Error("invalid equal key")
		return
	}

	if !newFilter.ParamItems.IsNotEqual("notequal") {
		t.Error("invalid not equal key")
		return
	}
	if newFilter.ParamItems.GetNotEqual("notequal") == nil {
		t.Error("invalid not equal key")
		return
	}

	if !newFilter.ParamItems.IsBelow("below") {
		t.Error("invalid below key")
		return
	}
	if newFilter.ParamItems.GetBelow("below") == nil {
		t.Error("invalid below key")
		return
	}

	if !newFilter.ParamItems.IsAbove("above") {
		t.Error("invalid above key")
		return
	}
	if newFilter.ParamItems.GetAbove("above") == nil {
		t.Error("invalid above key")
		return
	}

	if !newFilter.ParamItems.IsIn("in") {
		t.Error("invalid in key")
		return
	}
	if newFilter.ParamItems.GetIn("in") == nil {
		t.Error("invalid in key")
		return
	}

	if !newFilter.ParamItems.IsIn("inB") {
		t.Error("invalid in key")
		return
	}
	if newFilter.ParamItems.GetIn("inB") == nil {
		t.Error("invalid in key")
		return
	}

	if !newFilter.ParamItems.IsNotIn("notin") {
		t.Error("invalid not in key")
		return
	}
	if newFilter.ParamItems.GetNotIn("notin") == nil {
		t.Error("invalid not in key")
		return
	}

	if !newFilter.ParamItems.IsLike("like") {
		t.Error("invalid like key")
		return
	}
	if newFilter.ParamItems.GetLike("like") == nil {
		t.Error("invalid like key")
		return
	}
}
