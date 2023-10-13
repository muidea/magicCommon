package net

import "testing"

func TestJoinSuffix(t *testing.T) {
	valURL := "aa"
	suffix := "bb"
	ret := JoinSuffix(valURL, suffix)
	if ret != "aa/bb" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}

	valURL = "aa/"
	suffix = "bb"
	ret = JoinSuffix(valURL, suffix)
	if ret != "aa/bb" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}

	valURL = "/aa//"
	suffix = "bb"
	ret = JoinSuffix(valURL, suffix)
	if ret != "/aa/bb" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}
	valURL = "/aa/"
	suffix = "/bb"
	ret = JoinSuffix(valURL, suffix)
	if ret != "/aa/bb" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}
	valURL = "/aa/"
	suffix = "/bb/"
	ret = JoinSuffix(valURL, suffix)
	if ret != "/aa/bb/" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}

	valURL = "http://127.9.9.1/aa/?a=b"
	suffix = "/bb/"
	ret = JoinSuffix(valURL, suffix)
	if ret != "http://127.9.9.1/aa/bb/?a=b" {
		t.Error("JoinSuffix unexpect, ret:" + ret)
	}
}

func TestJoinPrefix(t *testing.T) {
	valURL := "aa"
	prefix := "bb"
	ret := JoinPrefix(valURL, prefix)
	if ret != "bb/aa" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}

	valURL = "aa/"
	prefix = "bb"
	ret = JoinPrefix(valURL, prefix)
	if ret != "bb/aa/" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}

	valURL = "/aa//"
	prefix = "bb"
	ret = JoinPrefix(valURL, prefix)
	if ret != "bb/aa/" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}
	valURL = "/aa/"
	prefix = "/bb"
	ret = JoinPrefix(valURL, prefix)
	if ret != "/bb/aa/" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}
	valURL = "/aa/"
	prefix = "/bb/"
	ret = JoinPrefix(valURL, prefix)
	if ret != "/bb/aa/" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}

	valURL = "http://127.9.9.1/aa/?a=b"
	prefix = "/bb/"
	ret = JoinPrefix(valURL, prefix)
	if ret != "http://127.9.9.1/bb/aa/?a=b" {
		t.Error("JoinPrefix unexpect, ret:" + ret)
	}
}

func TestParseRestAPIUrl(t *testing.T) {
	url := "/user/abc"
	dir, name := SplitRESTURL(url)
	if dir != "/user/" && name != "abc" {
		t.Errorf("SplitRESTURL failed, dir:%s,name:%s", dir, name)
	}

	url = "/user/abc/"
	dir, name = SplitRESTURL(url)
	if dir != "/user/abc/" && name != "" {
		t.Errorf("SplitRESTURL failed, dir:%s,name:%s", dir, name)
	}

	url = "/user/"
	dir, name = SplitRESTURL(url)
	if dir != "/user/" && name != "" {
		t.Errorf("SplitRESTURL failed, dir:%s,name:%s", dir, name)
	}
	url = "/user"
	dir, name = SplitRESTURL(url)
	if dir != "/" && name != "user" {
		t.Errorf("SplitRESTURL failed, dir:%s,name:%s", dir, name)
	}
}

func TestFormatRoutePattern(t *testing.T) {
	url := "/user/"
	id := "abc"
	pattern := FormatRoutePattern(url, id)
	if pattern != "/user/abc" {
		t.Errorf("FormatRoutePattern failed, url:%s, id:%s", url, id)
	}

	url = "/user/abc"
	id = "ef"
	pattern = FormatRoutePattern(url, id)
	if pattern != "/user/abc/ef" {
		t.Errorf("FormatRoutePattern failed, url:%s, id:%s", url, id)
	}

	url = "/user/abc"
	id = ""
	pattern = FormatRoutePattern(url, id)
	if pattern != "/user/abc/" {
		t.Errorf("FormatRoutePattern failed, url:%s, id:%s", url, id)
	}

	url = "/user/"
	id = ""
	pattern = FormatRoutePattern(url, id)
	if pattern != "/user/" {
		t.Errorf("FormatRoutePattern failed, url:%s, id:%s", url, id)
	}
}

func TestExtractID(t *testing.T) {
	url := "/abc/bcd/cde/"
	ret := ExtractID(url)
	if ret != "" {
		t.Errorf("ExtraceID failed, ret:%v", ret)
		return
	}

	url = "/abc/bcd/cde"
	ret = ExtractID(url)
	if ret != "cde" {
		t.Errorf("ExtraceID failed, ret:%v", ret)
		return
	}
}

func TestFormatID(t *testing.T) {
	url := "/abc/bcd/cde"
	ret := FormatID(url, 123)
	if ret != url {
		t.Errorf("FormatID failed,ret:%v", ret)
		return
	}

	url = "/abc/bcd/:id"
	ret = FormatID(url, 123)
	if ret != "/abc/bcd/123" {
		t.Errorf("FormatID failed,ret:%v", ret)
		return
	}
}

func TestSplitRESTID(t *testing.T) {
	url := "/abc/bcd/cde"
	id, err := SplitRESTID(url)
	if err == nil {
		t.Errorf("SpliteRESTID failed")
		return
	}

	url = "/abc/bcd/cde/"
	id, err = SplitRESTID(url)
	if err == nil {
		t.Errorf("SpliteRESTID failed")
		return
	}

	url = "/abc/bcd/cde/123"
	id, err = SplitRESTID(url)
	if err != nil {
		t.Errorf("SpliteRESTID failed")
		return
	}
	if id != 123 {
		t.Errorf("SpliteRESTID failed")
		return
	}
}

func TestSplitRESTPath(t *testing.T) {
	urlPath := "/abc/cde/efg"
	path, name := SplitRESTPath(urlPath)
	if path != "/abc/cde" || name != "efg" {
		t.Errorf("SplitRESTPath failed")
		return
	}

	urlPath = "/abc/cde/efg/"
	path, name = SplitRESTPath(urlPath)
	if path != "/abc/cde/efg" || name != "" {
		t.Errorf("SplitRESTPath failed")
	}

	urlPath = "abc/cde/efg/"
	path, name = SplitRESTPath(urlPath)
	if path != "abc/cde/efg" || name != "" {
		t.Errorf("SplitRESTPath failed")
	}

	urlPath = "abc/cde/efg"
	path, name = SplitRESTPath(urlPath)
	if path != "abc/cde" || name != "efg" {
		t.Errorf("SplitRESTPath failed")
	}
}
