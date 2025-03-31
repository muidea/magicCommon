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
