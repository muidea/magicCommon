package net

import (
	"fmt"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"
)

// JoinSuffix 合并Url路径后缀
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

// JoinPrefix 合并Url路径前缀
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

// SplitRESTURL 分割出RestAPI的路径和ID
func SplitRESTURL(url string) (string, string) {
	return path.Split(url)
}

func SplitRESTID(url string) (ret int, err error) {
	_, strID := SplitRESTURL(url)
	ret, err = strconv.Atoi(strID)
	return
}

func ExtractID(url string) (ret string) {
	_, ret = path.Split(url)
	return
}

func FormatID(url string, id interface{}) string {
	return strings.ReplaceAll(url, ":id", fmt.Sprintf("%v", id))
}

// FormatRoutePattern 格式化RoutePattern
func FormatRoutePattern(url string, id interface{}) string {
	if id != nil {
		return JoinSuffix(url, fmt.Sprintf("%v", id))
	}

	return path.Join(url, ":id")
}
