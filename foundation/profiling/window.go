// Package profiling provides opt-in, bounded process-window cost accounting.
// Labels are code-owned operations, never request values or SQL text.
package profiling

import (
	"encoding/json"
	"log/slog"
	"os"
	"sort"
	"sync"
	"time"
)

const maxKeys = 128

var bounds = [...]time.Duration{time.Millisecond, 5 * time.Millisecond, 10 * time.Millisecond, 25 * time.Millisecond, 50 * time.Millisecond, 100 * time.Millisecond, 250 * time.Millisecond, 500 * time.Millisecond, time.Second, 2 * time.Second}

type Entry struct {
	Category  string  `json:"category"`
	Operation string  `json:"operation"`
	Calls     uint64  `json:"calls"`
	Errors    uint64  `json:"errors"`
	TotalMS   float64 `json:"totalMs"`
	MaxMS     float64 `json:"maxMs"`
	// Buckets are disjoint; final bucket is greater than the last bound.
	Buckets [11]uint64 `json:"buckets"`
}
type Snapshot struct {
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
	BoundsMS  []float64 `json:"boundsMs"`
	Entries   []Entry   `json:"entries"`
}
type Window struct {
	mu      sync.Mutex
	started time.Time
	values  map[string]*Entry
}

func NewWindow(now time.Time) *Window { return &Window{started: now, values: make(map[string]*Entry)} }
func (w *Window) Record(category, operation string, duration time.Duration, failed bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	key := category + ":" + operation
	entry := w.values[key]
	if entry == nil {
		if len(w.values) >= maxKeys {
			category, operation, key = "overflow", "other", "overflow:other"
		}
		entry = w.values[key]
		if entry == nil {
			entry = &Entry{Category: category, Operation: operation}
			w.values[key] = entry
		}
	}
	if duration < 0 {
		duration = 0
	}
	ms := float64(duration) / float64(time.Millisecond)
	entry.Calls++
	if failed {
		entry.Errors++
	}
	entry.TotalMS += ms
	if ms > entry.MaxMS {
		entry.MaxMS = ms
	}
	bucket := len(bounds)
	for i, b := range bounds {
		if duration <= b {
			bucket = i
			break
		}
	}
	entry.Buckets[bucket]++
}
func (w *Window) Drain(now time.Time) Snapshot {
	w.mu.Lock()
	defer w.mu.Unlock()
	result := Snapshot{StartedAt: w.started, EndedAt: now, Entries: make([]Entry, 0, len(w.values))}
	for _, b := range bounds {
		result.BoundsMS = append(result.BoundsMS, float64(b)/float64(time.Millisecond))
	}
	for _, e := range w.values {
		result.Entries = append(result.Entries, *e)
	}
	sort.Slice(result.Entries, func(i, j int) bool {
		a, b := result.Entries[i], result.Entries[j]
		return a.Category+":"+a.Operation < b.Category+":"+b.Operation
	})
	w.values = make(map[string]*Entry)
	w.started = now
	return result
}

var processWindow *Window

func init() {
	value := os.Getenv("MAGIC_PROFILE_WINDOW")
	if value == "" {
		return
	}
	interval, err := time.ParseDuration(value)
	if err != nil || interval < time.Second || interval > 10*time.Minute {
		return
	}
	processWindow = NewWindow(time.Now().UTC())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for now := range ticker.C {
			snapshot := processWindow.Drain(now.UTC())
			if len(snapshot.Entries) > 0 {
				data, _ := json.Marshal(snapshot)
				slog.Info("performance window", "data", string(data))
			}
		}
	}()
}
func Enabled() bool { return processWindow != nil }
func Record(category, operation string, duration time.Duration, failed bool) {
	if processWindow != nil {
		processWindow.Record(category, operation, duration, failed)
	}
}
