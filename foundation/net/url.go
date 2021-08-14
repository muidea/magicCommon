package net

import (
	"fmt"
	"log"
	"net/url"
	"path"
)

// JoinSuffix 合并Url路径
func JoinSuffix(urlVal, suffix string) string {
	valURL, preErr := url.Parse(urlVal)
	if preErr != nil {
		log.Fatalf("illegal urlVal,preErr:%s", preErr.Error())
	}

	urlVal = valURL.Path
	if len(suffix) > 0 && suffix[len(suffix)-1] != '/' {
		urlVal = path.Join(urlVal, suffix)
	} else {
		urlVal = path.Join(urlVal, suffix) + "/"
	}
	valURL.Path = urlVal
	return valURL.String()
}

func JoinPrefix(urlVal, prefix string) string {
	valURL, preErr := url.Parse(urlVal)
	if preErr != nil {
		log.Fatalf("illegal urlVal,preErr:%s", preErr.Error())
	}

	urlVal = valURL.Path
	if len(urlVal) > 0 && urlVal[len(urlVal)-1] != '/' {
		urlVal = path.Join(prefix, urlVal)
	} else {
		urlVal = path.Join(prefix, urlVal) + "/"
	}

	valURL.Path = urlVal
	return valURL.String()
}

// SplitRESTAPI 分割出RestAPI的路径和ID
func SplitRESTAPI(url string) (string, string) {
	return path.Split(url)
}

// FormatRoutePattern 格式化RoutePattern
func FormatRoutePattern(url string, id interface{}) string {
	if id != nil {
		return JoinSuffix(url, fmt.Sprintf("%v", id))
	}

	return path.Join(url, ":id")
}
