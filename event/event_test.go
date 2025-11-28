package event

import (
	"context"
	"testing"

	cd "github.com/muidea/magicCommon/def"
)

func TestEventConstants(t *testing.T) {
	// Test that constants are defined correctly
	if innerDataKey != "_innerDataKey_" {
		t.Errorf("innerDataKey = %s, want %s", innerDataKey, "_innerDataKey_")
	}
	if innerValKey != "_innerValKey_" {
		t.Errorf("innerValKey = %s, want %s", innerValKey, "_innerValKey_")
	}

	// Test action constants
	if Action != "_action_" {
		t.Errorf("Action = %s, want %s", Action, "_action_")
	}
	if Add != "add" {
		t.Errorf("Add = %s, want %s", Add, "add")
	}
	if Del != "del" {
		t.Errorf("Del = %s, want %s", Del, "del")
	}
	if Mod != "mod" {
		t.Errorf("Mod = %s, want %s", Mod, "mod")
	}
	if Notify != "notify" {
		t.Errorf("Notify = %s, want %s", Notify, "notify")
	}
}

func TestNewValues(t *testing.T) {
	values := NewValues()
	if values == nil {
		t.Error("NewValues() returned nil")
	}
	if len(values) != 0 {
		t.Errorf("NewValues() returned non-empty map, got %d entries", len(values))
	}
}

func TestNewEventAndBaseEvent(t *testing.T) {
	id := "test/id"
	source := "test-source"
	destination := "test-destination"
	header := NewValues()
	header.Set("headerKey", "headerValue")
	data := "testData"

	event := NewEvent(id, source, destination, header, data)

	// Test basic properties
	if event.ID() != id {
		t.Errorf("event.ID() = %s, want %s", event.ID(), id)
	}
	if event.Source() != source {
		t.Errorf("event.Source() = %s, want %s", event.Source(), source)
	}
	if event.Destination() != destination {
		t.Errorf("event.Destination() = %s, want %s", event.Destination(), destination)
	}
	if event.Header().GetString("headerKey") != "headerValue" {
		t.Errorf("event.Header().GetString() = %s, want %s", event.Header().GetString("headerKey"), "headerValue")
	}
	if event.Data() != data {
		t.Errorf("event.Data() = %v, want %v", event.Data(), data)
	}

	// Test binding context
	ctx := context.Background()
	if event.Context() != nil {
		t.Errorf("event.Context() = %v, want nil", event.Context())
	}
	event.BindContext(ctx)
	if event.Context() != ctx {
		t.Errorf("event.Context() = %v, want %v", event.Context(), ctx)
	}

	// Test setting and getting data
	event.SetData("key1", 123)
	if val := event.GetData("key1"); val != 123 {
		t.Errorf("event.GetData() = %v, want %v", val, 123)
	}
	if val := event.GetData("nonexistent"); val != nil {
		t.Errorf("event.GetData() for nonexistent key = %v, want nil", val)
	}

	// Test Match method
	if !event.Match(id) {
		t.Errorf("event.Match(%s) = false, want true", id)
	}
	if !event.Match("test/+") {
		t.Errorf("event.Match(\"test/+\") = false, want true")
	}
	if event.Match("different/id") {
		t.Errorf("event.Match(\"different/id\") = true, want false")
	}
}

func TestNewEventWitchContext(t *testing.T) {
	id := "test/id"
	source := "test-source"
	destination := "test-destination"
	header := NewValues()
	ctx := context.Background()
	data := "testData"

	event := NewEventWitchContext(id, source, destination, header, ctx, data)

	if event.Context() != ctx {
		t.Errorf("event.Context() = %v, want %v", event.Context(), ctx)
	}
}

func TestNewResult(t *testing.T) {
	id := "test/id"
	source := "test-source"
	destination := "test-destination"

	result := NewResult(id, source, destination)

	// A new result should have an error
	if result.Error() == nil {
		t.Error("result.Error() = nil, want error")
	}

	// Test setting and getting data
	testData := "result data"
	customErr := cd.NewError(cd.Unexpected, "unknown message")
	result.Set(testData, customErr)

	data, err := result.Get()
	if data != testData {
		t.Errorf("result.Get() data = %v, want %v", data, testData)
	}
	if err != customErr {
		t.Errorf("result.Get() err = %v, want %v", err, customErr)
	}

	// Test setting and getting values
	result.SetVal("key1", 456)
	if val := result.GetVal("key1"); val != 456 {
		t.Errorf("result.GetVal() = %v, want %v", val, 456)
	}
	if val := result.GetVal("nonexistent"); val != nil {
		t.Errorf("result.GetVal() for nonexistent key = %v, want nil", val)
	}
}

func TestMatchValue(t *testing.T) {
	// Simple matches
	if !MatchValue("a", "a") {
		t.Error("MatchValue(\"a\", \"a\") = false, want true")
	}
	if MatchValue("a", "b") {
		t.Error("MatchValue(\"a\", \"b\") = true, want false")
	}
	if !MatchValue("a/b", "a/b") {
		t.Error("MatchValue(\"a/b\", \"a/b\") = false, want true")
	}

	// Plus wildcard
	if !MatchValue("a/+", "a/b") {
		t.Error("MatchValue(\"a/+\", \"a/b\") = false, want true")
	}
	if MatchValue("a/+", "a") {
		t.Error("MatchValue(\"a/+\", \"a\") = true, want false")
	}
	if !MatchValue("a/+/c", "a/b/c") {
		t.Error("MatchValue(\"a/+/c\", \"a/b/c\") = false, want true")
	}

	// ID wildcard
	if !MatchValue("a/:id", "a/123") {
		t.Error("MatchValue(\"a/:id\", \"a/123\") = false, want true")
	}
	if MatchValue("a/:id", "a") {
		t.Error("MatchValue(\"a/:id\", \"a\") = true, want false")
	}

	// Hash wildcard
	if !MatchValue("a/#", "a/b") {
		t.Error("MatchValue(\"a/#\", \"a/b\") = false, want true")
	}
	if !MatchValue("a/#", "a/b/c") {
		t.Error("MatchValue(\"a/#\", \"a/b/c\") = false, want true")
	}
	if MatchValue("a/#", "a") {
		t.Error("MatchValue(\"a/#\", \"a\") = true, want false")
	}
	if !MatchValue("a/b/#", "a/b/c/d") {
		t.Error("MatchValue(\"a/b/#\", \"a/b/c/d\") = false, want true")
	}

	// Complex patterns
	if !MatchValue("a/+/c/#", "a/b/c/d") {
		t.Error("MatchValue(\"a/+/c/#\", \"a/b/c/d\") = false, want true")
	}
	if MatchValue("a/+/x/#", "a/b/c/d") {
		t.Error("MatchValue(\"a/+/x/#\", \"a/b/c/d\") = true, want false")
	}
	if !MatchValue("a/#/c", "a/b/c") {
		t.Error("MatchValue(\"a/#/c\", \"a/b/c\") = false, want true")
	}
	if !MatchValue("a/#/c", "a/b/x/c") {
		t.Error("MatchValue(\"a/#/c\", \"a/b/x/c\") = false, want true")
	}
}

func TestGetAs(t *testing.T) {
	// 创建一个测试用的 Result
	result := NewResult("test/id", "test-source", "test-destination")

	// 测试成功转换字符串类型
	t.Run("successful string conversion", func(t *testing.T) {
		testData := "hello world"
		result.Set(testData, nil)

		val, err := GetAs[string](result)
		if err != nil {
			t.Errorf("GetAs[string] returned error: %v", err)
		}
		if val != testData {
			t.Errorf("GetAs[string] = %v, want %v", val, testData)
		}
	})

	// 测试成功转换整数类型
	t.Run("successful int conversion", func(t *testing.T) {
		testData := 42
		result.Set(testData, nil)

		val, err := GetAs[int](result)
		if err != nil {
			t.Errorf("GetAs[int] returned error: %v", err)
		}
		if val != testData {
			t.Errorf("GetAs[int] = %v, want %v", val, testData)
		}
	})

	// 测试成功转换结构体类型
	t.Run("successful struct conversion", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		testData := TestStruct{Name: "John", Age: 30}
		result.Set(testData, nil)

		val, err := GetAs[TestStruct](result)
		if err != nil {
			t.Errorf("GetAs[TestStruct] returned error: %v", err)
		}
		if val != testData {
			t.Errorf("GetAs[TestStruct] = %v, want %v", val, testData)
		}
	})

	// 测试类型不匹配的情况
	t.Run("type mismatch", func(t *testing.T) {
		testData := "not an int"
		result.Set(testData, nil)

		val, err := GetAs[int](result)
		if err == nil {
			t.Error("GetAs[int] should return error for string input")
		} else if err.Code != cd.Unexpected {
			t.Errorf("Error code = %v, want %v", err.Code, cd.Unexpected)
		}
		// 当err不为nil时，val应该是该类型的零值
		var zeroVal int
		if val != zeroVal {
			t.Errorf("GetAs[int] should return zero value for type mismatch, got %v", val)
		}
	})

	// 测试 nil 值的情况
	t.Run("nil value", func(t *testing.T) {
		result.Set(nil, nil)

		val, err := GetAs[string](result)
		if err != nil {
			t.Errorf("GetAs[string] should not return error for nil value, got %v", err)
		}
		if val != "" {
			t.Errorf("GetAs[string] should return empty string for nil value, got %v", val)
		}
	})

	// 测试带有错误的情况
	t.Run("with error", func(t *testing.T) {
		testData := "test data"
		customErr := cd.NewError(cd.Unexpected, "custom error")
		result.Set(testData, customErr)

		val, err := GetAs[string](result)
		if err != customErr {
			t.Errorf("GetAs should preserve original error, got %v, want %v", err, customErr)
		}
		if val != testData {
			t.Errorf("GetAs should return correct value, got %v, want %v", val, testData)
		}
	})
}

func TestGetValAs(t *testing.T) {
	// 创建一个测试用的 Result
	result := NewResult("test/id", "test-source", "test-destination")

	// 测试成功转换字符串类型
	t.Run("successful string conversion", func(t *testing.T) {
		testData := "hello world"
		result.SetVal("stringKey", testData)

		val, ok := GetValAs[string](result, "stringKey")
		if !ok {
			t.Error("GetValAs[string] should return ok = true")
		}
		if val != testData {
			t.Errorf("GetValAs[string] = %v, want %v", val, testData)
		}
	})

	// 测试成功转换整数类型
	t.Run("successful int conversion", func(t *testing.T) {
		testData := 42
		result.SetVal("intKey", testData)

		val, ok := GetValAs[int](result, "intKey")
		if !ok {
			t.Error("GetValAs[int] should return ok = true")
		}
		if val != testData {
			t.Errorf("GetValAs[int] = %v, want %v", val, testData)
		}
	})

	// 测试成功转换结构体类型
	t.Run("successful struct conversion", func(t *testing.T) {
		type TestStruct struct {
			Name string
			Age  int
		}
		testData := TestStruct{Name: "John", Age: 30}
		result.SetVal("structKey", testData)

		val, ok := GetValAs[TestStruct](result, "structKey")
		if !ok {
			t.Error("GetValAs[TestStruct] should return ok = true")
		}
		if val != testData {
			t.Errorf("GetValAs[TestStruct] = %v, want %v", val, testData)
		}
	})

	// 测试类型不匹配的情况
	t.Run("type mismatch", func(t *testing.T) {
		testData := "not an int"
		result.SetVal("mismatchKey", testData)

		val, ok := GetValAs[int](result, "mismatchKey")
		if ok {
			t.Error("GetValAs[int] should return ok = false for type mismatch")
		}
		if val != 0 {
			t.Errorf("GetValAs[int] should return zero value for type mismatch, got %v", val)
		}
	})

	// 测试 nil 值的情况
	t.Run("nil value", func(t *testing.T) {
		result.SetVal("nilKey", nil)

		val, ok := GetValAs[string](result, "nilKey")
		if ok {
			t.Error("GetValAs should return ok = false for nil value")
		}
		if val != "" {
			t.Errorf("GetValAs should return zero value for nil value, got %v", val)
		}
	})

	// 测试不存在的键
	t.Run("non-existent key", func(t *testing.T) {
		val, ok := GetValAs[string](result, "nonexistent")
		if ok {
			t.Error("GetValAs should return ok = false for non-existent key")
		}
		if val != "" {
			t.Errorf("GetValAs should return zero value for non-existent key, got %v", val)
		}
	})

	// 测试多个键值对
	t.Run("multiple key-value pairs", func(t *testing.T) {
		stringData := "string value"
		intData := 123
		result.SetVal("multiString", stringData)
		result.SetVal("multiInt", intData)

		// 测试获取字符串值
		strVal, strOk := GetValAs[string](result, "multiString")
		if !strOk || strVal != stringData {
			t.Errorf("GetValAs for multiString failed: ok=%v, val=%v", strOk, strVal)
		}

		// 测试获取整数值
		intVal, intOk := GetValAs[int](result, "multiInt")
		if !intOk || intVal != intData {
			t.Errorf("GetValAs for multiInt failed: ok=%v, val=%v", intOk, intVal)
		}
	})
}
