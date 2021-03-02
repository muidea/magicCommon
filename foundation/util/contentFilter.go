package util

import "net/http"

type ContentFilter struct {
	ItemVal map[string]string
}

func NewContentFilter(items []string) *ContentFilter {
	itemVal := map[string]string{}
	for _, val := range items {
		itemVal[val] = ""
	}

	return &ContentFilter{ItemVal: itemVal}
}

func (s *ContentFilter) Decode(request *http.Request) {
	newItemVal := map[string]string{}
	for key := range s.ItemVal {
		val := request.URL.Query().Get(key)
		newItemVal[key] = val
	}
}
