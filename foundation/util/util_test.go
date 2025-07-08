package util

import "testing"

func TestExtractTelephone(t *testing.T) {
	val := "贾夕夕  2.22  13808099541"
	telephone := ExtractTelephone(val)
	if telephone != "13808099541" {
		t.Errorf("extrace telephone failed,raw:%s, telephone:%s", val, telephone)
		return
	}

	// 3.2213918686045
	val = "jenny  2.18  /2.22"
	telephone = ExtractTelephone(val)
	if telephone != "" {
		t.Errorf("extrace telephone failed,raw:%s", val)
		return
	}
	val = "3.2213918686045"
	telephone = ExtractTelephone(val)
	if telephone == "" {
		t.Errorf("extrace telephone failed,raw:%s", val)
		return
	}
}

func TestStringSet(t *testing.T) {
	set := StringSet{}
	set = set.Add("a")
	set = set.Add("b")
	set = set.Add("c")
	if len(set) != 3 {
		t.Errorf("add string failed, expect 3, but %d", len(set))
		return
	}

	set = set.Add("a")
	if len(set) != 3 {
		t.Errorf("add string failed, expect 3, but %d", len(set))
		return
	}

	if set[0] != "a" || set[1] != "b" || set[2] != "c" {
		t.Errorf("add string failed, expect a,b,c, but %s,%s,%s", set[0], set[1], set[2])
		return
	}

	set = set.Add("d")
	if len(set) != 4 {
		t.Errorf("add string failed, expect 4, but %d", len(set))
		return
	}
}
