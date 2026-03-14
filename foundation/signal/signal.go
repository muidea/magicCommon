package signal

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"log/slog"
)

type signalEntry struct {
	mu     sync.RWMutex
	ch     chan interface{}
	closed bool
}

func newSignalEntry() *signalEntry {
	return &signalEntry{ch: make(chan interface{}, 1)}
}

func (s *signalEntry) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	close(s.ch)
	s.closed = true
}

func (s *signalEntry) send(val interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.closed {
		return errors.New("signal already closed")
	}

	s.ch <- val
	return nil
}

type Gard struct {
	signalChanMap sync.Map
}

func (s *Gard) PutSignal(id int) (err error) {
	_, ok := s.signalChanMap.Load(id)
	if ok {
		msg := fmt.Sprintf("duplicate signal %d", id)
		err = errors.New(msg)
		slog.Error(msg)
		return
	}

	s.signalChanMap.Store(id, newSignalEntry())
	return
}

func (s *Gard) CleanSignal(id int) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		return
	}

	s.signalChanMap.Delete(id)
	signalChan.(*signalEntry).close()
}

func (s *Gard) WaitSignal(id, timeOut int) (ret interface{}, err error) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = errors.New(msg)
		slog.Error(msg)
		return
	}
	entry := signalChan.(*signalEntry)

	defer func() {
		s.signalChanMap.Delete(id)
		entry.close()
	}()

	if timeOut < 0 {
		timeOut = 60 * 60
	}
	timeOutVal := time.Duration(timeOut) * time.Second
	select {
	case val, ok := <-entry.ch:
		if ok {
			ret = val
		}
	case <-time.After(timeOutVal):
		msg := fmt.Sprintf("wait signal %d timeout", id)
		err = errors.New(msg)
		slog.Warn(msg)
	}
	return
}

func (s *Gard) TriggerSignal(id int, val interface{}) (err error) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = errors.New(msg)
		slog.Error(msg)
		return
	}

	err = signalChan.(*signalEntry).send(val)
	if err != nil {
		slog.Error("trigger signal failed", "id", id, "error", err)
	}
	return
}

func (s *Gard) Reset() {
	s.signalChanMap.Range(func(key, value any) bool {
		s.signalChanMap.Delete(key)
		value.(*signalEntry).close()
		return true
	})
}
