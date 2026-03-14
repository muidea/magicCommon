package event

import "testing"

// TestNewHubWithOptions 验证 NewHubWithOptions 不改变语义
func TestNewHubWithOptions(t *testing.T) {
	hub := NewHubWithOptions(10,
		WithPerDestinationChanSize(32),
		WithHubActionChanSize(64),
		WithWorkerPoolSize(16),
	)
	defer hub.Terminate()

	handler := &eventHandler{handlerID: "/opt-handler"}
	hub.Subscribe("/opt-event", handler)

	ev := NewEvent("/opt-event", "/", handler.ID(), NewValues(), "data")
	result := hub.Send(ev)

	val, err := result.Get()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "data" {
		t.Fatalf("unexpected result data: %v", val)
	}
	if !handler.handled {
		t.Fatalf("handler should be called")
	}
}
