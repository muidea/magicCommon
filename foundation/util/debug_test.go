package util

import (
	"bytes"
	"testing"
)

func TestGetStack(t *testing.T) {
	// 调用 GetStack 函数，跳过当前和调用 GetStack 的帧
	stackTrace := GetStack(1)

	// 检查返回的堆栈跟踪是否不为空
	if len(stackTrace) == 0 {
		t.Error("Expected non-empty stack trace, got empty")
	}

	// 检查堆栈跟踪是否包含一些基本的堆栈信息，如文件名和行号
	if !bytes.Contains(stackTrace, []byte("debug_test.go")) {
		t.Errorf("Expected stack trace to contain 'debug_test.go', got %s", stackTrace)
	}
}
