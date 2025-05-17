package task

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
	"github.com/stretchr/testify/assert"
)

func calcOffset(intervalValue, offsetValue time.Duration) time.Duration {
	return func() time.Duration {
		now := time.Now()
		log.Infof("%v", now)
		nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
		if intervalValue < 24*time.Hour {
			return (nowOffset/intervalValue+1)*intervalValue - nowOffset
		}

		return (offsetValue + intervalValue - nowOffset + 24*time.Hour) % (24 * time.Hour)
	}()
}

func TestBackgroundRoutine_Timer(t *testing.T) {
	//intervalValue := 10 * time.Minute
	intervalValue := 24 * time.Hour
	//offsetValue := time.Duration(0)
	offsetValue := 1 * time.Hour

	curOffset := calcOffset(intervalValue, offsetValue)
	log.Infof("%v", curOffset)

	intervalValue = 10 * time.Minute
	offsetValue = 0
	curOffset = calcOffset(intervalValue, offsetValue)
	log.Infof("%v", curOffset)
}

type asyncTask struct {
	wg          *sync.WaitGroup
	taskRoutine BackgroundRoutine
	index       int
}

func (s *asyncTask) Run() {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond * 10)
	fmt.Printf("task index:%d\n", s.index)
	if s.taskRoutine != nil {
		s.wg.Add(1)
		s.taskRoutine.AsyncTask(&subTask{
			wg: s.wg,
		})
	}

	s.wg.Done()
}

type subTask struct {
	wg *sync.WaitGroup
}

func (s *subTask) Run() {
	fmt.Printf("subTask running!\n")
	s.wg.Done()
}

func TestNewBackgroundRoutine(t *testing.T) {
	wg := &sync.WaitGroup{}
	taskRoutine := NewBackgroundRoutine(300)

	idx := 0
	for ; idx < 10; idx++ {
		wg.Add(1)
		taskRoutine.SyncTask(&asyncTask{
			wg:          wg,
			index:       idx,
			taskRoutine: taskRoutine,
		})
	}

	for ; idx < 1000; idx++ {
		wg.Add(1)
		taskRoutine.AsyncTask(&asyncTask{
			wg:    wg,
			index: idx,
		})
	}

	wg.Wait()
}

type timerTask struct {
	timerCount int
}

func (s *timerTask) Run() {
	s.timerCount++
}

func TestTimer(t *testing.T) {
	timerTaskPtr := &timerTask{}
	taskRoutine := NewBackgroundRoutine(300)
	taskRoutine.Timer(timerTaskPtr, 1*time.Second, 0)

	time.Sleep(10 * time.Second)
	assert.True(t, timerTaskPtr.timerCount > 8, "timerTaskPtr.timerCount>8")
}
