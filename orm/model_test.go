package orm

import "testing"

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int8 `json:"id"`
	// Name 名称
	Name      string  `json:"name"`
	Value     float32 `json:"value"`
	TimeStamp int     `json:"timeStamp"`
}

func TestModel(t *testing.T) {
	info := getModelInfo(&Unit{})
	if info == nil {
		t.Errorf("getModelInfo failed,")
		return
	}

	info.Dump()
}
