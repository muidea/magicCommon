package event

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

func TestMatchID(t *testing.T) {
	cases := []struct {
		filter, topic string
		want          bool
	}{
		{"device/+", "device/123", true}, {"device/+", "device/", true},
		{"device/+", "device", false}, {"device/#", "device", true},
		{"device/#", "device/", true}, {"device/#", "device/a/b", true},
		{"device/#/status", "device/a/status", false}, {"device/:id", "device/42", false},
		{"device/:id", "device/:id", true}, {"#", "$SYS/status", false},
		{"+/status", "$SYS/status", false}, {"$SYS/#", "$SYS/status", true},
		{"device//status", "device//status", true}, {"/", "/", true},
		{"a/+", "a/b/c", false}, {"a/b", "A/b", false}, {"a+", "a+", false},
		{"a/+x", "a/x", false}, {"a/#", "a/#", false}, {"+/+", "/finance", true},
		{"#", "/", true}, {"+", "/", false}, {"a/+/#", "a/", true},
		{"a/+/#", "a", false}, {"$SYS/#", "$SYS", true}, {"+/+", "/", true},
		{"/+/status", "/$SYS/status", true}, {"é", "e\u0301", false},
	}
	for _, tc := range cases {
		t.Run(tc.filter+"_"+tc.topic, func(t *testing.T) {
			if got := MatchValue(tc.filter, tc.topic); got != tc.want {
				t.Fatalf("MatchValue=%v want %v", got, tc.want)
			}
			if got := NewEvent(tc.topic, "", "", nil, nil).Match(tc.filter); got != tc.want {
				t.Fatalf("Event.Match=%v", got)
			}
			parsed, err := ParseTopicFilter(tc.filter)
			if err == nil && parsed.MatchTopic(tc.topic) != tc.want {
				t.Fatal("compiled matcher disagrees")
			}
		})
	}
}

func TestTopicValidationBoundaries(t *testing.T) {
	for _, valid := range []string{"/", "a//", ":id", " ", "\ufeffx", strings.Repeat("a", maxTopicBytes), strings.Repeat("/", maxTopicBytes), strings.Repeat("界", maxTopicBytes/3)} {
		if err := ValidateTopicName(valid); err != nil {
			t.Fatal(err)
		}
		if err := ValidateTopicFilter(valid); err != nil {
			t.Fatal(err)
		}
	}
	for _, invalid := range []string{"", "a\x00b", string([]byte{255}), strings.Repeat("a", maxTopicBytes+1), strings.Repeat("界", maxTopicBytes/3) + "x"} {
		for _, validate := range []func(string) *cd.Error{ValidateTopicName, ValidateTopicFilter} {
			if err := validate(invalid); err == nil || err.Code != cd.IllegalParam {
				t.Fatalf("invalid string accepted (%d bytes): %v", len(invalid), err)
			}
		}
	}
	for _, filter := range []string{"a+", "a/#/b", "a/+x", "a/b#", "##", "a/#/", "+a"} {
		if err := ValidateTopicFilter(filter); err == nil {
			t.Fatalf("accepted %q", filter)
		}
	}
	for _, topic := range []string{"+", "#", "a/+", "a/#"} {
		if ValidateTopicName(topic) == nil {
			t.Fatalf("accepted name %q", topic)
		}
	}
	if !MatchValue(strings.Repeat("+/", 32767)+"+", strings.Repeat("/", 32767)) {
		t.Fatal("maximum-depth filter did not match empty levels")
	}
	if (TopicFilter{}).MatchTopic("valid") {
		t.Fatal("zero filter matched")
	}
}

func TestOverlappingSubscriptionsSelectOneHandler(t *testing.T) {
	for _, tc := range []struct {
		filters     []string
		topic, want string
	}{
		{[]string{"a", "a/#"}, "a", "a"},
		{[]string{"a/b", "a/+"}, "a/b", "a/b"},
		{[]string{"a/+", "a/#"}, "a/b", "a/+"},
		{[]string{"a/+/c", "a/b/#"}, "a/b/c", "a/b/#"},
	} {
		for _, reverse := range []bool{false, true} {
			hub := NewHub(8).(DrainingHub)
			observer := NewSimpleObserver("owner", hub)
			calls := map[string]int{}
			for index := range tc.filters {
				if reverse {
					index = len(tc.filters) - 1 - index
				}
				filter := tc.filters[index]
				if err := observer.Subscribe(filter, func(_ Event, r Result) {
					calls[filter]++
					if r != nil {
						r.Set(filter, nil)
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
			for i := 0; i < 8; i++ {
				r := hub.Send(NewEvent(tc.topic, "caller", "owner", nil, nil))
				value, err := r.Get()
				if err != nil || value != tc.want {
					t.Fatalf("selected %v, %v; want %s", value, err, tc.want)
				}
			}
			if err := hub.PostChecked(NewEvent(tc.topic, "caller", "owner", nil, nil)); err != nil {
				t.Fatal(err)
			}
			if err := hub.Drain(context.Background()); err != nil {
				t.Fatal(err)
			}
			if calls[tc.want] != 9 || len(calls) != 1 {
				t.Fatalf("duplicate or wrong callbacks: %v", calls)
			}
			if err := observer.Unsubscribe(tc.want); err != nil {
				t.Fatal(err)
			}
			r := hub.Send(NewEvent(tc.topic, "caller", "owner", nil, nil))
			value, err := r.Get()
			if err != nil || value == tc.want {
				t.Fatalf("remaining handler unavailable: %v, %v", value, err)
			}
			for _, filter := range tc.filters {
				if err := observer.Unsubscribe(filter); err != nil {
					t.Fatal(err)
				}
			}
			if hub.Send(NewEvent(tc.topic, "caller", "owner", nil, nil)).Error() == nil {
				t.Fatal("all subscriptions removed but still matched")
			}
			hub.Terminate(context.Background())
		}
	}
}

type topicObserver struct{ id, matchID string }

func (o *topicObserver) ID() string      { return o.id }
func (o *topicObserver) MatchID() string { return o.matchID }
func (*topicObserver) Notify(_ Event, r Result) {
	if r != nil {
		r.Set(true, nil)
	}
}

type uncomparableObserver struct{ values []string }

func (uncomparableObserver) ID() string           { return "uncomparable" }
func (uncomparableObserver) Notify(Event, Result) {}

type nestedValueObserver struct{ payload any }

func (nestedValueObserver) ID() string { return "nested-value" }
func (nestedValueObserver) Notify(_ Event, result Result) {
	if result != nil {
		result.Set(true, nil)
	}
}

func TestObserverIdentityChecksNestedInterfaceValues(t *testing.T) {
	for _, tc := range []struct {
		name       string
		payload    any
		comparable bool
	}{
		{name: "slice", payload: []string{"x"}},
		{name: "map", payload: map[string]int{"x": 1}},
		{name: "nested-array", payload: [1]any{[]int{1}}},
		{name: "nested-struct", payload: struct{ Value any }{Value: []int{1}}},
		{name: "comparable", payload: "stable", comparable: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hub := NewHub(8).(*hubImpl)
			defer hub.Terminate(context.Background())
			observer := nestedValueObserver{payload: tc.payload}
			if sameObserver(observer, observer) != tc.comparable {
				t.Fatal("incorrect value comparability")
			}
			if err := hub.Subscribe("one", observer); err != nil {
				t.Fatal(err)
			}
			if err := hub.Send(NewEvent("one", "caller", observer.ID(), nil, nil)).Error(); err != nil {
				t.Fatal(err)
			}
			err := hub.Subscribe("two", observer)
			if tc.comparable {
				if err != nil {
					t.Fatal("same comparable value could not add a filter", err)
				}
				if err := hub.Unsubscribe("two", observer); err != nil {
					t.Fatal(err)
				}
				// Warm the same cache after the successful unsubscribe.
				if err := hub.Send(NewEvent("one", "caller", observer.ID(), nil, nil)).Error(); err != nil {
					t.Fatal(err)
				}
			} else if err == nil || err.Code != cd.Duplicated {
				t.Fatalf("unprovable identity must be Duplicated: %v", err)
			}
			if len(hub.event2Observer) != 1 || len(hub.topicFilters) != 1 ||
				len(hub.observerRecords) != 1 || hub.observerRecords[observer.ID()].filters != 1 || len(hub.eventMatchCache) != 1 {
				t.Fatal("identity check changed registrations or cache")
			}
			if err := hub.Send(NewEvent("one", "caller", observer.ID(), nil, nil)).Error(); err != nil {
				t.Fatal("original registration was damaged", err)
			}
			if err := hub.Unsubscribe("one", observer); err != nil {
				t.Fatal(err)
			}
			if err := hub.Subscribe("two", &topicObserver{id: observer.ID()}); err != nil {
				t.Fatal("identity was not released", err)
			}
		})
	}
}

func TestTopicSubscriptionRejectionHasNoEffects(t *testing.T) {
	hub := NewHub(8).(*hubImpl)
	defer hub.Terminate(context.Background())
	first := &topicObserver{id: "owner", matchID: "owner"}
	if err := hub.Subscribe("valid", first); err != nil {
		t.Fatal(err)
	}
	if hub.Send(NewEvent("valid", "caller", "owner", nil, nil)).Error() != nil {
		t.Fatal("missing initial handler")
	}
	for _, filter := range []string{"a/#/b", "a+", "", string([]byte{255})} {
		if err := hub.Subscribe(filter, first); err == nil || err.Code != cd.IllegalParam {
			t.Fatalf("invalid subscribe: %v", err)
		}
	}
	if err := hub.Subscribe("new", &topicObserver{id: "owner", matchID: "owner"}); err == nil || err.Code != cd.Duplicated {
		t.Fatalf("identity collision: %v", err)
	}
	if err := hub.Subscribe("new", &topicObserver{id: "owner", matchID: "different"}); err == nil || err.Code != cd.Duplicated {
		t.Fatalf("another instance must be an identity conflict: %v", err)
	}
	first.matchID = "changed"
	if err := hub.Subscribe("new", first); err == nil || err.Code != cd.IllegalParam {
		t.Fatalf("changed matcher: %v", err)
	}
	first.matchID = "owner"
	if len(hub.topicFilters) != 1 || len(hub.event2Observer) != 1 || len(hub.eventMatchCache) != 1 {
		t.Fatal("rejection mutated registry or cache")
	}
	if err := hub.Subscribe("valid", first); err != nil {
		t.Fatal(err)
	}
	value := uncomparableObserver{values: []string{"x"}}
	if err := hub.Subscribe("one", value); err != nil {
		t.Fatal(err)
	}
	if err := hub.Subscribe("two", value); err == nil || err.Code != cd.Duplicated {
		t.Fatalf("uncomparable identity conflict: %v", err)
	}
	if err := hub.Unsubscribe("one", value); err != nil {
		t.Fatal(err)
	}
}

type customTopicEvent struct{ Event }

func (customTopicEvent) Match(string) bool { return false }

func TestHubRoutesByTopicRatherThanCustomEventMatch(t *testing.T) {
	hub := NewHub(4)
	defer hub.Terminate(context.Background())
	observer := NewSimpleObserver("owner", hub)
	if err := observer.Subscribe("a/+", func(_ Event, r Result) { r.Set("ok", nil) }); err != nil {
		t.Fatal(err)
	}
	r := hub.Send(customTopicEvent{NewEvent("a/b", "caller", "owner", nil, nil)})
	if value, err := r.Get(); value != "ok" || err != nil {
		t.Fatalf("custom matcher changed routing: %v %v", value, err)
	}
}

func TestInvalidPublicationsDoNotEnterHub(t *testing.T) {
	hub := NewHub(4).(*hubImpl)
	defer hub.Terminate(context.Background())
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	defer slog.SetDefault(previous)
	for _, ev := range []Event{nil, NewEvent("", "", "lane", nil, nil), NewEvent("a/+", "", "lane", nil, "secret-data"), customTopicEvent{NewEvent("bad\x00secret", "", "lane", nil, nil)}} {
		if err := hub.PostChecked(ev); err == nil || err.Code != cd.IllegalParam {
			t.Fatalf("PostChecked invalid error: %v", err)
		}
		hub.Post(ev)
		if r := hub.Send(ev); r == nil || r.Error() == nil || r.Error().Code != cd.IllegalParam {
			t.Fatalf("Send invalid result: %v", r)
		}
	}
	if len(hub.laneKey2ActionChannel) != 0 || len(hub.eventMatchCache) != 0 || hub.operations != 0 {
		t.Fatal("invalid event changed hub state")
	}
	if strings.Count(logs.String(), "event publish rejected") != 4 || strings.Contains(logs.String(), "secret") {
		t.Fatal("invalid input logged or missing rejection")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := hub.PostChecked(NewEventWithContext("valid", "", "lane", nil, ctx, nil)); err == nil || err.Code != cd.Timeout {
		t.Fatalf("canceled post: %v", err)
	}
	if err := hub.TerminateChecked(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := hub.PostChecked(NewEvent("valid", "", "lane", nil, nil)); err == nil || err.Code != cd.InvalidOperation {
		t.Fatalf("stopped post: %v", err)
	}
}

func TestDestinationContractRemainsIndependent(t *testing.T) {
	for _, tc := range []struct {
		destination, match string
		want               bool
	}{
		{"/modules/a", "/modules/a", true}, {"/#", "/modules/a", true}, {"/#", "owner", false},
		{"#", "owner", true}, {"", "owner", false}, {"a/:id", "a/42", true}, {"a/42", "a/:id", true},
		{"a/+", "a/", false}, {"a/#", "a", false}, {"a/#/b", "a/x/b", true}, {"#", "$SYS/status", true}, {"/#", "/", false}, {"", "", true},
	} {
		if got := matchDestination(tc.destination, tc.match); got != tc.want {
			t.Fatalf("address %q %q=%v", tc.destination, tc.match, got)
		}
	}
}

func TestPostCheckedRejectsBusyLaneWithoutLateDelivery(t *testing.T) {
	hub := NewHubWithOptions(1, WithPerLaneChanSize(1)).(*hubImpl)
	entered, release := make(chan struct{}), make(chan struct{})
	var delivered atomic.Int64
	observer := NewSimpleObserver("owner", hub)
	if err := observer.Subscribe("valid", func(ev Event, _ Result) {
		delivered.Add(1)
		if ev.Data() == "block" {
			close(entered)
			<-release
		}
	}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { hub.Terminate(context.Background()) })
	defer close(release)
	if err := hub.PostChecked(NewEvent("valid", "", "owner", nil, "block")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-entered:
	case <-time.After(time.Second):
		t.Fatal("handler did not enter")
	}
	if err := hub.PostChecked(NewEvent("valid", "", "owner", nil, "queued")); err != nil {
		t.Fatal(err)
	}
	if err := hub.PostChecked(NewEvent("valid", "", "owner", nil, "rejected")); err == nil || err.Code != cd.Timeout {
		t.Fatalf("full lane: %v", err)
	}
	if err := hub.PostChecked(NewEvent("invalid/+", "", "owner", nil, "invalid")); err == nil || err.Code != cd.IllegalParam {
		t.Fatalf("invalid topic lost to capacity error: %v", err)
	}
	// Cleanup drains after releasing the handler. Only the two admitted events
	// may execute; a rejected enqueue cannot arrive later.
	t.Cleanup(func() {
		if err := hub.Drain(context.Background()); err != nil {
			t.Fatal(err)
		}
		if got := delivered.Load(); got != 2 {
			t.Fatalf("delivered %d, want 2", got)
		}
	})
}
