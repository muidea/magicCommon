package def

import (
	"encoding/json"
	"testing"
)

func TestCommonResult(t *testing.T) {
	type Value struct {
		AInt   int     `json:"aInt"`
		BFloat float64 `json:"BFloat"`
	}

	type Compose struct {
		Result
		CValue Value `json:"value"`
	}

	compse := &Compose{Result: NewResult(), CValue: Value{AInt: 123, BFloat: 234.567}}
	byteVal, _ := json.Marshal(compse)

	commonResult := &CommonResult{}
	_ = json.Unmarshal(byteVal, commonResult)

	if commonResult.Error == nil {
		t.Errorf("encode failed")
		return
	}

	value := &Value{}
	_ = json.Unmarshal(commonResult.Value, value)
	if value.AInt != compse.CValue.AInt {
		t.Errorf("unmarshal failed")
	}
}
