package util

import (
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

	req := &http.Request{URL: &url.URL{RawQuery: val.Encode()}}

	newFilter := NewFilter()
	newFilter.Decode(req)
}
