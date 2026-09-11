package common

import (
	"context"
	"fmt"
	cd "github.com/muidea/magicCommon/def"
)

// ShutdownStarter must only close admission/cancel work, without releasing
// dependencies or waiting. Every owner is notified before any owner is drained.
type ShutdownStarter interface{ BeginShutdown(context.Context) }

// Quiescer retains dependencies on failure and must support retry. All guards
// must succeed before the process calls any Teardown method.
type Quiescer interface {
	Quiesce(context.Context) *cd.Error
}

func (s *PluginMgr) shutdownSnapshot() []any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	values := s.entityList
	if s.setupStarted {
		values = s.activeList
	}
	var active []any
	for _, value := range values {
		if !s.released[value] {
			active = append(active, value)
		}
	}
	return active
}

func (s *PluginMgr) BeginShutdown(ctx context.Context) *cd.Error {
	var first *cd.Error
	for _, owner := range s.shutdownSnapshot() {
		if starter, ok := owner.(ShutdownStarter); ok {
			err := guardShutdown(func() *cd.Error { starter.BeginShutdown(ctx); return nil })
			if first == nil {
				first = err
			}
		}
	}
	return first
}

func (s *PluginMgr) Quiesce(ctx context.Context) *cd.Error {
	for _, owner := range s.shutdownSnapshot() {
		if guard, ok := owner.(Quiescer); ok {
			if err := guardShutdown(func() *cd.Error { return guard.Quiesce(ctx) }); err != nil {
				return err
			}
		}
	}
	return nil
}

func guardShutdown(fn func() *cd.Error) (err *cd.Error) {
	defer func() {
		if value := recover(); value != nil {
			if cause, ok := value.(*cd.Error); ok {
				err = cause
				return
			}
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("shutdown guard panicked: %v", value))
		}
	}()
	return fn()
}
