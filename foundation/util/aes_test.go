package util

import (
	"testing"
)

func TestEncryptByAes(t *testing.T) {
	raw := "hey worldfsdfsdfsdfdfsfunc (s *BaseClient) GetContextValues() url.Values {"
	key := "123"

	encryVal, encryErr := EncryptByAes(raw, key)
	if encryErr != nil {
		t.Errorf("EncryptByAes failed, err:%s", encryErr.Error())
		return
	}

	t.Logf("%s\n", encryVal)
	rawVal, rawErr := DecryptByAes(encryVal, key)
	if rawErr != nil {
		t.Errorf("DecryptByAes failed, err:%s", rawErr.Error())
		return
	}
	if rawVal != raw {
		t.Errorf("DecryptByAes failed")
		return
	}
}
