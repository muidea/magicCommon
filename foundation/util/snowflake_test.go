package util

import "testing"

func TestNewSnowflakeNode(t *testing.T) {
	node, err := NewSnowflakeNode(1)
	if err != nil {
		t.Errorf("new snowflake node failed, err:%s", err.Error())
		return
	}

	id := node.Generate()
	t.Logf("id:%d", id)
}
