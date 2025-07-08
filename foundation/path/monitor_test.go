package path

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockObserver struct {
	events []Event
	mu     sync.Mutex
}

func (m *mockObserver) OnEvent(event Event) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *mockObserver) getEvents() []Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

func TestMonitor_Basic(t *testing.T) {
	monitor, err := NewMonitor(nil)
	assert.NoError(t, err)
	assert.NotNil(t, monitor)

	err = monitor.Start()
	assert.NoError(t, err)
	defer monitor.Stop()

	observer := &mockObserver{}
	monitor.AddObserver(observer)

	tmpDir := t.TempDir()

	err = monitor.AddPath(tmpDir)
	assert.NoError(t, err)

	// Test create file
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for event processing

	events := observer.getEvents()
	assert.Equal(t, testFile, events[0].Path)
	assert.Equal(t, Create, events[0].Op)

	// Test modify file
	err = os.WriteFile(testFile, []byte("modified"), 0644)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for event processing

	events = observer.getEvents()
	assert.GreaterOrEqual(t, len(events), 2)
	assert.Equal(t, testFile, events[1].Path)
	assert.Equal(t, Modify, events[1].Op)

	// Test remove file
	err = os.Remove(testFile)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second) // Wait for event processing

	events = observer.getEvents()
	assert.GreaterOrEqual(t, len(events), 3)
	assert.Equal(t, testFile, events[2].Path)
	assert.Equal(t, Remove, events[2].Op)
}

func TestMonitor_Ignore(t *testing.T) {
	monitor, err := NewMonitor([]string{"ignore.txt"})
	assert.NoError(t, err)

	err = monitor.Start()
	assert.NoError(t, err)
	defer monitor.Stop()

	observer := &mockObserver{}
	monitor.AddObserver(observer)

	tmpDir := t.TempDir()
	err = monitor.AddPath(tmpDir)
	assert.NoError(t, err)

	// Create ignored file
	ignoreFile := filepath.Join(tmpDir, "ignore.txt")
	err = os.WriteFile(ignoreFile, []byte("test"), 0644)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Should not receive event for ignored file
	assert.Empty(t, observer.getEvents())
}

func TestMonitor_RemoveObserver(t *testing.T) {
	monitor, err := NewMonitor(nil)
	assert.NoError(t, err)

	err = monitor.Start()
	assert.NoError(t, err)
	defer monitor.Stop()

	observer := &mockObserver{}
	monitor.AddObserver(observer)

	tmpDir := t.TempDir()
	err = monitor.AddPath(tmpDir)
	assert.NoError(t, err)

	// Remove observer before file operation
	monitor.RemoveObserver(observer)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// Should not receive event after removal
	assert.Empty(t, observer.getEvents())
}

func TestMonitor_RemovePath(t *testing.T) {
	monitor, err := NewMonitor(nil)
	assert.NoError(t, err)

	err = monitor.Start()
	assert.NoError(t, err)
	defer monitor.Stop()

	observer := &mockObserver{}
	monitor.AddObserver(observer)

	tmpDir := t.TempDir()
	err = monitor.AddPath(tmpDir)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Remove path before file operation
	err = monitor.RemovePath(tmpDir)
	assert.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)

	time.Sleep(1 * time.Second)

	// Should not receive event after path removal
	assert.Empty(t, observer.getEvents())
}

func TestMonitor_Concurrent(t *testing.T) {
	monitor, err := NewMonitor(nil)
	assert.NoError(t, err)

	err = monitor.Start()
	assert.NoError(t, err)
	defer monitor.Stop()

	var wg sync.WaitGroup
	observers := make([]*mockObserver, 10)

	// Add multiple observers concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			observer := &mockObserver{}
			monitor.AddObserver(observer)
			observers[idx] = observer
		}(i)
	}
	wg.Wait()

	tmpDir := t.TempDir()
	err = monitor.AddPath(tmpDir)
	assert.NoError(t, err)

	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	assert.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// All observers should receive the event
	for _, observer := range observers {
		assert.NotNil(t, observer)
		events := observer.getEvents()
		assert.GreaterOrEqual(t, len(events), 1)
		assert.Equal(t, testFile, events[0].Path)
	}
}
