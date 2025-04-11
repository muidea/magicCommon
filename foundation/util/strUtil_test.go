package util

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
)

type Car struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Desc string `json:"desc"`
	Age  int    `json:"age"`
	Car  *Car   `json:"car"`
}

func TestIntArray2Str(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		expected string
	}{
		{"Non-empty array", []int{1, 2}, "1,2"},
		{"Empty array", []int{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IntArray2Str(tt.input)
			if result != tt.expected {
				t.Errorf("IntArray2Str failed, got: %s, want: %s", result, tt.expected)
			}
		})
	}
}

func TestStr2IntArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []int
		ok       bool
	}{
		{"Empty string", "", []int{}, true},
		{"Single number", "1", []int{1}, true},
		{"Leading comma", ",1", []int{1}, true},
		{"Trailing comma", "1,", []int{1}, true},
		{"Both leading and trailing comma", ",1,", []int{1}, true},
		{"Multiple numbers", ",1,2,3,4", []int{1, 2, 3, 4}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := Str2IntArray(tt.input)
			if !ok || !compareIntSlices(result, tt.expected) {
				t.Errorf("Str2IntArray failed, got: %v, want: %v, ok: %v", result, tt.expected, ok)
			}
		})
	}
}

func TestMarshalString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"Integer", 1234, "1234"},
		{"String", "1234", "1234"},
		{"Float", 12.34, "12.34"},
		{"Bool false", false, "false"},
		{"Bool true", true, "true"},
		{"Complex string", "61d383cb134f4db6a367046ffac3051d", "61d383cb134f4db6a367046ffac3051d"},
		{"User object", &User{ID: 110, Name: "Hello", Desc: "hey boy", Age: 123, Car: &Car{ID: 100, Name: "Car"}}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MarshalString(tt.input)
			if tt.name == "User object" {
				if result == "" {
					t.Errorf("MarshalString failed for User object, got empty string")
				}
				log.Infof(result)
			} else if result != tt.expected {
				t.Errorf("MarshalString failed, got: %s, want: %s", result, tt.expected)
			}
		})
	}
}

func TestUnmarshalString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"Integer", "1234", float64(1234)},
		{"String", "a1234", "a1234"},
		{"Float", "12.34", float64(12.34)},
		{"Bool false", "false", false},
		{"Bool true", "true", true},
		{"Complex string", "61d383cb134f4db6a367046ffac3051d", "61d383cb134f4db6a367046ffac3051d"},
		{"JSON object", `{"id":110,"name":"Hello","desc":"hey boy","age":123,"car":{"id":100,"name":"Car"}}`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UnmarshalString(tt.input)
			if tt.name == "JSON object" {
				if result == nil {
					t.Errorf("UnmarshalString failed for JSON object, got nil")
				}
			} else if result != tt.expected {
				t.Errorf("UnmarshalString failed, got: %v, want: %v", result, tt.expected)
			}
		})
	}
}

func TestExtractSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Multi-line content", "This is a test content.\nThis is the second line.", "This is a test content."},
		{"Single line content", "Single line content", "Single line content"},
		{"Leading newlines", "\n\n\nThis is a test content.\nThis is the second line.", "This is a test content."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractSummary(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractSummary failed, got: %s, want: %s", result, tt.expected)
			}
		})
	}
}

func TestCleanStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Leading and trailing commas", ", 1, 2, 3, ", "1, 2, 3"},
		{"Trailing comma", "1, 2, 3,", "1, 2, 3"},
		{"Leading and trailing commas with spaces", ",1, 2, 3,", "1, 2, 3"},
		{"Spaces only", "  1, 2, 3  ", "1, 2, 3"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanStr(tt.input)
			if result != tt.expected {
				t.Errorf("cleanStr failed, got: %s, want: %s", result, tt.expected)
			}
		})
	}
}

// Helper function to compare two integer slices
func compareIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}