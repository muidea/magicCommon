package signal

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"log/slog"
)

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

	signalChan := make(chan interface{}, 1)
	s.signalChanMap.Store(id, signalChan)
	return
}

func (s *Gard) CleanSignal(id int) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			slog.Error("clean signal id unexpected, err:errInfo", "field", id, "error", errInfo)
		}
	}()

	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		return
	}

	s.signalChanMap.Delete(id)
	close(signalChan.(chan interface{}))
}

func (s *Gard) WaitSignal(id, timeOut int) (ret interface{}, err error) {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			slog.Error("wait signal id unexpected, err:errInfo", "field", id, "error", errInfo)
		}
	}()

	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = errors.New(msg)
		slog.Error(msg)
		return
	}
	defer func() {
		s.signalChanMap.Delete(id)
		close(signalChan.(chan interface{}))
	}()

	if timeOut < 0 {
		timeOut = 60 * 60
	}
	timeOutVal := time.Duration(timeOut) * time.Second
	select {
	case val, ok := <-signalChan.(chan interface{}):
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
	defer func() {
		if errInfo := recover(); errInfo != nil {
			slog.Error("trigger signal id unexpected, err:errInfo", "field", id, "error", errInfo)
		}
	}()

	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = errors.New(msg)
		slog.Error(msg)
		return
	}

	signalChan.(chan interface{}) <- val
	return
}

func (s *Gard) Reset() {
	defer func() {
		if errInfo := recover(); errInfo != nil {
			slog.Error("reset signal chan map unexpected, err:%v", errInfo)
		}
	}()

	s.signalChanMap.Range(func(key, value any) bool {
		s.signalChanMap.Delete(key)
		close(value.(chan interface{}))
		return true
	})

}
