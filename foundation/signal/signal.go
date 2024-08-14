package signal

import (
	"fmt"
	"sync"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
)

type Gard struct {
	signalChanMap sync.Map
}

func (s *Gard) PutSignal(id int) (err error) {
	_, ok := s.signalChanMap.Load(id)
	if ok {
		msg := fmt.Sprintf("duplicate signal %d", id)
		err = fmt.Errorf(msg)
		log.Errorf(msg)
		return
	}

	signalChan := make(chan bool, 1)
	s.signalChanMap.Store(id, signalChan)
	return
}

func (s *Gard) CleanSignal(id int) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		return
	}

	s.signalChanMap.Delete(id)
	close(signalChan.(chan bool))
}

func (s *Gard) WaitSignal(id, timeOut int) (err error) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = fmt.Errorf(msg)
		log.Errorf(msg)
		return
	}

	if timeOut < 0 {
		timeOut = 60 * 60
	}
	timeOutVal := time.Duration(timeOut) * time.Second
	select {
	case <-signalChan.(chan bool):
	case <-time.After(timeOutVal):
		msg := fmt.Sprintf("wait signal %d timeout", id)
		err = fmt.Errorf(msg)
		log.Warnf(msg)
		signalChan.(chan bool) <- true
	}

	close(signalChan.(chan bool))
	s.signalChanMap.Delete(id)
	return
}

func (s *Gard) TriggerSignal(id int) (err error) {
	signalChan, signalOK := s.signalChanMap.Load(id)
	if !signalOK {
		msg := fmt.Sprintf("can't find signal %d", id)
		err = fmt.Errorf(msg)
		log.Errorf(msg)
		return
	}

	signalChan.(chan bool) <- true
	return
}
