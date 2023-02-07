package util

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"testing"
)

func TestContentFilter_Encode(t *testing.T) {
	filter := NewFilter()
	filter.Set("name", "123")
	filter.Like("desc", "hello world")
	filter.Below("age", 40)

	val := url.Values{}
	val = filter.Encode(val)

	log.Print(val.Encode())

	req := &http.Request{URL: &url.URL{RawQuery: val.Encode()}}

	newFilter := NewFilter()
	newFilter.Decode(req)

	ii := 100
	byteVal, _ := json.Marshal(ii)
	log.Print(string(byteVal))

	array := []float32{12.23, 23.45}
	byteVal, _ = json.Marshal(array)
	log.Print(string(byteVal))
}
