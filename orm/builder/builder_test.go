package builder

import "testing"

// Unit 单元信息
type Unit struct {
	//ID 唯一标示单元
	ID int `json:"id" orm:"id key auto"`
	// Name 名称
	Name  string  `json:"name" orm:"name"`
	Value float32 `json:"value" orm:"value"`
}

func TestBuilder(t *testing.T) {
	obj := &Unit{ID: 10, Name: "Hello world", Value: 12.3456}

	builder := NewBuilder(obj)

	builder.BuildCreateSchema()

	builder.BuildDropSchema()

	builder.BuildInsert()

	builder.BuildUpdate()

	builder.BuildDelete()

	builder.BuildQuery()
}
