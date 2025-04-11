package util

import "testing"

func TestNewSnowFlakeNode(t *testing.T) {
	node, err := NewSnowFlakeNode(1)
	if err != nil {
		t.Errorf("new snowflake node failed, err:%s", err.Error())
		return
	}

	id := node.Generate()
	t.Logf("id:%d", id)
}
