package util

import (
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

// TestConvertValue_BasicTypes 测试基本类型转换
func TestConvertValue_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		ok       bool
	}{
		// 字符串类型
		{
			name:     "string to string",
			input:    "hello",
			expected: "hello",
			ok:       true,
		},
		// 整数类型
		{
			name:     "int to int",
			input:    42,
			expected: 42,
			ok:       true,
		},
		{
			name:     "int64 to int64",
			input:    int64(100),
			expected: int64(100),
			ok:       true,
		},
		// 浮点数类型
		{
			name:     "float64 to float64",
			input:    3.14,
			expected: 3.14,
			ok:       true,
		},
		// 布尔类型
		{
			name:     "bool to bool",
			input:    true,
			expected: true,
			ok:       true,
		},
		{
			name:     "bool false to bool",
			input:    false,
			expected: false,
			ok:       true,
		},
		// nil 值
		{
			name:     "nil value",
			input:    nil,
			expected: 0,
			ok:       false,
		},
		// 类型不匹配（应该失败）
		{
			name:     "int to string should fail",
			input:    123,
			expected: "",
			ok:       false,
		},
		{
			name:     "string to int should fail",
			input:    "123",
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case string:
				got, ok := ConvertValue[string](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[string](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[string](%v) = %v, want %v", tt.input, got, expected)
				}
			case int:
				got, ok := ConvertValue[int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int](%v) = %v, want %v", tt.input, got, expected)
				}
			case int64:
				got, ok := ConvertValue[int64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int64](%v) = %v, want %v", tt.input, got, expected)
				}
			case float64:
				got, ok := ConvertValue[float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[float64](%v) = %v, want %v", tt.input, got, expected)
				}
			case bool:
				got, ok := ConvertValue[bool](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[bool](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[bool](%v) = %v, want %v", tt.input, got, expected)
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}

// TestConvertValue_JsonNumber 测试 json.Number 类型转换
func TestConvertValue_JsonNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    json.Number
		expected any
		ok       bool
	}{
		{
			name:     "json.Number to int",
			input:    json.Number("42"),
			expected: 42,
			ok:       true,
		},
		{
			name:     "json.Number to int64",
			input:    json.Number("100"),
			expected: int64(100),
			ok:       true,
		},
		{
			name:     "json.Number to float64",
			input:    json.Number("3.14"),
			expected: 3.14,
			ok:       true,
		},
		{
			name:     "json.Number with decimal to int (truncation)",
			input:    json.Number("3.99"),
			expected: 3,
			ok:       true,
		},
		{
			name:     "json.Number negative to int",
			input:    json.Number("-10"),
			expected: -10,
			ok:       true,
		},
		{
			name:     "json.Number out of range for int8",
			input:    json.Number("1000"),
			expected: int8(0),
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case int:
				got, ok := ConvertValue[int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int](%v) = %v, want %v", tt.input, got, expected)
				}
			case int64:
				got, ok := ConvertValue[int64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int64](%v) = %v, want %v", tt.input, got, expected)
				}
			case float64:
				got, ok := ConvertValue[float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[float64](%v) = %v, want %v", tt.input, got, expected)
				}
			case int8:
				got, ok := ConvertValue[int8](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int8](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int8](%v) = %v, want %v", tt.input, got, expected)
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}

// TestConvertValue_NumericConversions 测试数字类型之间的转换
func TestConvertValue_NumericConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		ok       bool
	}{
		// int 到其他整数类型
		{
			name:     "int to int8 (in range)",
			input:    100,
			expected: int8(100),
			ok:       true,
		},
		{
			name:     "int to int8 (out of range)",
			input:    200,
			expected: int8(0),
			ok:       false,
		},
		{
			name:     "int to int16",
			input:    30000,
			expected: int16(30000),
			ok:       true,
		},
		{
			name:     "int to int32",
			input:    1000000,
			expected: int32(1000000),
			ok:       true,
		},
		{
			name:     "int to int64",
			input:    5000000000,
			expected: int64(5000000000),
			ok:       true,
		},
		// int 到无符号整数
		{
			name:     "int to uint (positive)",
			input:    100,
			expected: uint(100),
			ok:       true,
		},
		{
			name:     "int to uint (negative should fail)",
			input:    -10,
			expected: uint(0),
			ok:       false,
		},
		{
			name:     "int to uint8",
			input:    200,
			expected: uint8(200),
			ok:       true,
		},
		{
			name:     "int to uint16",
			input:    40000,
			expected: uint16(40000),
			ok:       true,
		},
		// float 到整数（截断）
		{
			name:     "float64 to int (truncation)",
			input:    3.99,
			expected: 3,
			ok:       true,
		},
		{
			name:     "float64 to int8",
			input:    127.0,
			expected: int8(127),
			ok:       true,
		},
		{
			name:     "float64 to int8 (out of range)",
			input:    128.0,
			expected: int8(0),
			ok:       false,
		},
		// 整数到浮点数
		{
			name:     "int to float64",
			input:    42,
			expected: 42.0,
			ok:       true,
		},
		{
			name:     "int to float32",
			input:    100,
			expected: float32(100.0),
			ok:       true,
		},
		// 无符号整数到有符号整数
		{
			name:     "uint to int",
			input:    uint(100),
			expected: 100,
			ok:       true,
		},
		{
			name:     "uint8 to int8 (in range)",
			input:    uint8(100),
			expected: int8(100),
			ok:       true,
		},
		{
			name:     "uint8 to int8 (out of range)",
			input:    uint8(200),
			expected: int8(0),
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case int:
				got, ok := ConvertValue[int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int](%v) = %v, want %v", tt.input, got, expected)
				}
			case int8:
				got, ok := ConvertValue[int8](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int8](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int8](%v) = %v, want %v", tt.input, got, expected)
				}
			case int16:
				got, ok := ConvertValue[int16](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int16](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int16](%v) = %v, want %v", tt.input, got, expected)
				}
			case int32:
				got, ok := ConvertValue[int32](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int32](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int32](%v) = %v, want %v", tt.input, got, expected)
				}
			case int64:
				got, ok := ConvertValue[int64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int64](%v) = %v, want %v", tt.input, got, expected)
				}
			case uint:
				got, ok := ConvertValue[uint](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[uint](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[uint](%v) = %v, want %v", tt.input, got, expected)
				}
			case uint8:
				got, ok := ConvertValue[uint8](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[uint8](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[uint8](%v) = %v, want %v", tt.input, got, expected)
				}
			case uint16:
				got, ok := ConvertValue[uint16](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[uint16](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[uint16](%v) = %v, want %v", tt.input, got, expected)
				}
			case float32:
				got, ok := ConvertValue[float32](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[float32](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[float32](%v) = %v, want %v", tt.input, got, expected)
				}
			case float64:
				got, ok := ConvertValue[float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[float64](%v) = %v, want %v", tt.input, got, expected)
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}

// TestConvertValue_SliceConversions 测试切片类型转换
func TestConvertValue_SliceConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		ok       bool
	}{
		// []any 到具体类型切片
		{
			name:     "[]any to []string",
			input:    []any{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
			ok:       true,
		},
		{
			name:     "[]any to []int",
			input:    []any{1, 2, 3},
			expected: []int{1, 2, 3},
			ok:       true,
		},
		{
			name:     "[]any to []float64",
			input:    []any{1.1, 2.2, 3.3},
			expected: []float64{1.1, 2.2, 3.3},
			ok:       true,
		},
		{
			name:     "[]any to []bool",
			input:    []any{true, false, true},
			expected: []bool{true, false, true},
			ok:       true,
		},
		// 具体类型切片到 []any
		{
			name:     "[]int to []any",
			input:    []int{10, 20, 30},
			expected: []any{10, 20, 30},
			ok:       true,
		},
		{
			name:     "[]string to []any",
			input:    []string{"x", "y", "z"},
			expected: []any{"x", "y", "z"},
			ok:       true,
		},
		// 数字类型切片转换
		{
			name:     "[]int to []int8",
			input:    []int{100, 200, 300},
			expected: []int8{100, -56, 44}, // 200超出int8范围，300超出int8范围
			ok:       false,
		},
		{
			name:     "[]float64 to []int (truncation)",
			input:    []float64{1.1, 2.9, 3.5},
			expected: []int{1, 2, 3},
			ok:       true,
		},
		// 空切片
		{
			name:     "empty []any to []string",
			input:    []any{},
			expected: []string{},
			ok:       true,
		},
		{
			name:     "empty []int to []any",
			input:    []int{},
			expected: []any{},
			ok:       true,
		},
		// 包含 json.Number 的切片
		{
			name:     "[]json.Number to []int",
			input:    []json.Number{json.Number("1"), json.Number("2"), json.Number("3")},
			expected: []int{1, 2, 3},
			ok:       true,
		},
		{
			name:     "[]json.Number to []float64",
			input:    []json.Number{json.Number("1.5"), json.Number("2.7")},
			expected: []float64{1.5, 2.7},
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case []string:
				got, ok := ConvertValue[[]string](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]string](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]string](%v) = %v, want %v", tt.input, got, expected)
				}
			case []int:
				got, ok := ConvertValue[[]int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]int](%v) = %v, want %v", tt.input, got, expected)
				}
			case []float64:
				got, ok := ConvertValue[[]float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]float64](%v) = %v, want %v", tt.input, got, expected)
				}
			case []bool:
				got, ok := ConvertValue[[]bool](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]bool](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]bool](%v) = %v, want %v", tt.input, got, expected)
				}
			case []any:
				got, ok := ConvertValue[[]any](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]any](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]any](%v) = %v, want %v", tt.input, got, expected)
				}
			case []int8:
				got, ok := ConvertValue[[]int8](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[[]int8](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[[]int8](%v) = %v, want %v", tt.input, got, expected)
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}

// TestConvertValue_MapConversions 测试映射类型转换
func TestConvertValue_MapConversions(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected any
		ok       bool
	}{
		// map[string]any 到 map[string]string
		{
			name:     "map[string]any to map[string]string",
			input:    map[string]any{"a": "hello", "b": "world"},
			expected: map[string]string{"a": "hello", "b": "world"},
			ok:       true,
		},
		{
			name:     "map[string]any to map[string]int",
			input:    map[string]any{"x": 1, "y": 2},
			expected: map[string]int{"x": 1, "y": 2},
			ok:       true,
		},
		// map[string]any 到 map[string]float64
		{
			name:     "map[string]any to map[string]float64",
			input:    map[string]any{"a": 1.5, "b": 2.7},
			expected: map[string]float64{"a": 1.5, "b": 2.7},
			ok:       true,
		},
		// 具体类型映射到 map[string]any
		{
			name:     "map[string]int to map[string]any",
			input:    map[string]int{"x": 10, "y": 20},
			expected: map[string]any{"x": float64(10), "y": float64(20)}, // JSON 会将数字转换为 float64
			ok:       true,
		},
		{
			name:     "map[string]string to map[string]any",
			input:    map[string]string{"key": "value"},
			expected: map[string]any{"key": "value"},
			ok:       true,
		},
		// 空映射
		{
			name:     "empty map[string]any to map[string]string",
			input:    map[string]any{},
			expected: map[string]string{},
			ok:       true,
		},
		{
			name:     "empty map[string]int to map[string]any",
			input:    map[string]int{},
			expected: map[string]any{},
			ok:       true,
		},
		// 包含 json.Number 的映射
		{
			name:     "map[string]json.Number to map[string]int",
			input:    map[string]json.Number{"a": json.Number("1"), "b": json.Number("2")},
			expected: map[string]int{"a": 1, "b": 2},
			ok:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch expected := tt.expected.(type) {
			case map[string]string:
				got, ok := ConvertValue[map[string]string](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[map[string]string](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[map[string]string](%v) = %v, want %v", tt.input, got, expected)
				}
			case map[string]int:
				got, ok := ConvertValue[map[string]int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[map[string]int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[map[string]int](%v) = %v, want %v", tt.input, got, expected)
				}
			case map[string]float64:
				got, ok := ConvertValue[map[string]float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[map[string]float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && !reflect.DeepEqual(got, expected) {
					t.Errorf("ConvertValue[map[string]float64](%v) = %v, want %v", tt.input, got, expected)
				}
			case map[string]any:
				got, ok := ConvertValue[map[string]any](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[map[string]any](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok {
					// 对于 map[string]any，我们需要比较内容而不是类型
					// 因为 reflect.DeepEqual 会检查值的具体类型
					if len(got) != len(expected) {
						t.Errorf("ConvertValue[map[string]any](%v) length = %v, want %v", tt.input, len(got), len(expected))
					} else {
						for k, v := range expected {
							gv, exists := got[k]
							if !exists {
								t.Errorf("ConvertValue[map[string]any](%v) missing key %v", tt.input, k)
							} else if !reflect.DeepEqual(gv, v) {
								t.Errorf("ConvertValue[map[string]any](%v)[%v] = %v (%T), want %v (%T)", tt.input, k, gv, gv, v, v)
							}
						}
					}
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}

// TestConvertValue_EdgeCases 测试边界情况
func TestConvertValue_EdgeCases(t *testing.T) {
	// 检查 int 的大小（32位还是64位）
	const is64Bit = ^uint(0)>>63 == 1

	tests := []struct {
		name     string
		input    any
		expected any
		ok       bool
		skipOn64 bool // 在64位系统上跳过这个测试
	}{
		// 零值转换
		{
			name:     "zero int to int",
			input:    0,
			expected: 0,
			ok:       true,
		},
		{
			name:     "zero float64 to float64",
			input:    0.0,
			expected: 0.0,
			ok:       true,
		},
		// 大数字转换
		{
			name:     "large int64 to int (should fail if out of range)",
			input:    int64(1<<63 - 1), // MaxInt64
			expected: 0,
			ok:       false,
			skipOn64: true, // 在64位系统上跳过，因为转换会成功
		},
		{
			name:     "large float64 to int (should fail if out of range)",
			input:    1e100,
			expected: 0,
			ok:       false,
		},
		// 特殊浮点数值
		{
			name:     "NaN to int (should fail)",
			input:    math.NaN(),
			expected: 0,
			ok:       false,
		},
		{
			name:     "Inf to int (should fail)",
			input:    math.Inf(1),
			expected: 0,
			ok:       false,
		},
		// 指针类型（应该失败）
		{
			name:     "pointer to int (should fail)",
			input:    new(int),
			expected: 0,
			ok:       false,
		},
		// 结构体类型（应该失败）
		{
			name:     "struct to int (should fail)",
			input:    struct{}{},
			expected: 0,
			ok:       false,
		},
		// 通道类型（应该失败）
		{
			name:     "channel to int (should fail)",
			input:    make(chan int),
			expected: 0,
			ok:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 如果在64位系统上需要跳过这个测试
			if tt.skipOn64 && is64Bit {
				t.Skip("Skipping test on 64-bit system")
			}

			switch expected := tt.expected.(type) {
			case int:
				got, ok := ConvertValue[int](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[int](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[int](%v) = %v, want %v", tt.input, got, expected)
				}
			case float64:
				got, ok := ConvertValue[float64](tt.input)
				if ok != tt.ok {
					t.Errorf("ConvertValue[float64](%v) ok = %v, want %v", tt.input, ok, tt.ok)
				}
				if ok && got != expected {
					t.Errorf("ConvertValue[float64](%v) = %v, want %v", tt.input, got, expected)
				}
			default:
				t.Errorf("unexpected expected type: %T", expected)
			}
		})
	}
}
