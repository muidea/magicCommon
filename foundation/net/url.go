package net

import (
	"log"
	"net/url"
	"path"
)

// JoinURL 合并Url路径
func JoinURL(prefix, subfix string) string {
	preURL, preErr := url.Parse(prefix)
	if preErr != nil {
		log.Fatalf("illegal prefix,preErr:%s", preErr.Error())
	}

	prefix = preURL.Path
	if len(subfix) > 0 && subfix[len(subfix)-1] != '/' {
		prefix = path.Join(prefix, subfix)
	} else {
		prefix = path.Join(prefix, subfix) + "/"
	}
	preURL.Path = prefix
	return preURL.String()
}

// SplitRESTAPI 分割出RestAPI的路径和ID
func SplitRESTAPI(url string) (string, string) {
	return path.Split(url)
}

// FormatRoutePattern 格式化RoutePattern
func FormatRoutePattern(url, id string) string {
	if len(id) == 0 {
		return JoinURL(url, "")
	}

	return path.Join(url, ":id")
}
