package os

import "testing"

func TestGetSystemMemory(t *testing.T) {
	memoerySize, memoryErr := GetSystemMemory()
	t.Log(memoerySize, memoryErr)
}
