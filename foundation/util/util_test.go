package util

import "testing"

func TestExtractTelephone(t *testing.T) {
	val := "贾夕夕  2.22  13808099541"
	telephone := ExtractTelephone(val)
	if telephone != "13808099541" {
		t.Errorf("extrace telephone failed,raw:%s, telephone:%s", val, telephone)
		return
	}

	val = "jenny  2.18  /2.22"
	telephone = ExtractTelephone(val)
	if telephone != "" {
		t.Errorf("extrace telephone failed,raw:%s", val)
		return
	}
}
