package util

import "testing"

func TestNewUUID(t *testing.T) {
	id := NewUUID()

	t.Logf("id:%s", id)
}
