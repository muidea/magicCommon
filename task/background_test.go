package task

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func calcOffset(intervalValue, offsetValue time.Duration) time.Duration {
	return func() time.Duration {
		now := time.Now()
		log.Printf("%v", now)
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
	log.Printf("%v", curOffset)

	intervalValue = 10 * time.Minute
	offsetValue = 0
	curOffset = calcOffset(intervalValue, offsetValue)
	log.Printf("%v", curOffset)
}

type asyncTask struct {
	wg    *sync.WaitGroup
	index int
}

func (s *asyncTask) Run() {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond * 10)
	fmt.Printf("task index:%d\n", s.index)
	s.wg.Done()
}

func TestNewBackgroundRoutine(t *testing.T) {
	wg := &sync.WaitGroup{}
	taskRoutine := NewBackgroundRoutine(300)

	idx := 0
	for ; idx < 10; idx++ {
		wg.Add(1)
		taskRoutine.SyncTask(&asyncTask{
			wg:    wg,
			index: idx,
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
