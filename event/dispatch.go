package event

import (
	"context"

	cd "github.com/muidea/magicCommon/def"
)

func (s *hubImpl) activeLanes(ctx context.Context) []string {
	if ctx == nil {
		return nil
	}
	frame, _ := ctx.Value(laneExecutionContextKey{}).(*laneExecutionFrame)
	var keys []string
	seen := map[string]bool{}
	for ; frame != nil; frame = frame.parent {
		if frame.hub != s || !frame.active.Load() {
			break
		}
		if !seen[frame.key] {
			keys = append(keys, frame.key)
			seen[frame.key] = true
		}
	}
	return keys
}

// Account for accepted work until dispatch really completes, not merely until
// a canceled sender stops waiting. This also pins dependencies for nested Send.
func (s *hubImpl) beginOperation(ctx context.Context) bool {
	s.operationsMu.Lock()
	defer s.operationsMu.Unlock()
	if s.terminateFlag.Load() && (s.operations == 0 || len(s.activeLanes(ctx)) == 0) {
		return false
	}
	if s.operations == 0 {
		s.idle = make(chan struct{})
	}
	s.operations++
	return true
}

func (s *hubImpl) endOperation() {
	s.operationsMu.Lock()
	defer s.operationsMu.Unlock()
	s.operations--
	if s.operations == 0 {
		close(s.idle)
	}
}

func (s *hubImpl) Drain(ctx context.Context) *cd.Error {
	if ctx == nil {
		ctx = context.Background()
	}
	s.operationsMu.Lock()
	idle := s.idle
	s.operationsMu.Unlock()
	select {
	case <-idle:
		return nil
	default:
	}
	select {
	case <-idle:
		return nil
	case <-ctx.Done():
		return cd.NewError(cd.Timeout, "hub dispatches are still running")
	}
}

// Same-chain ancestry permits safe synchronous reentry; independent chains may
// not enter each other's occupied lanes. Reject a cyclic wait before enqueueing
// instead of deadlocking or concurrently running handlers in a serialized lane.
func (s *hubImpl) beginLaneWait(ev Event, target string) (func(), *cd.Error) {
	ancestors := s.activeLanes(ev.Context())
	if len(ancestors) == 0 {
		return func() {}, nil
	}
	s.waitsMu.Lock()
	defer s.waitsMu.Unlock()
	for _, source := range ancestors {
		if s.hasLanePath(target, source, map[string]bool{}) {
			return nil, cd.NewError(cd.InvalidOperation, "cyclic synchronous event lane dependency")
		}
	}
	if s.waits == nil {
		s.waits = map[string]map[string]int{}
	}
	for _, source := range ancestors {
		if s.waits[source] == nil {
			s.waits[source] = map[string]int{}
		}
		s.waits[source][target]++
	}
	return func() {
		s.waitsMu.Lock()
		defer s.waitsMu.Unlock()
		for _, source := range ancestors {
			s.waits[source][target]--
			if s.waits[source][target] == 0 {
				delete(s.waits[source], target)
			}
			if len(s.waits[source]) == 0 {
				delete(s.waits, source)
			}
		}
	}, nil
}

func (s *hubImpl) hasLanePath(from, to string, visited map[string]bool) bool {
	if from == to {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true
	for next := range s.waits[from] {
		if s.hasLanePath(next, to, visited) {
			return true
		}
	}
	return false
}
