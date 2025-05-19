package task

import (
	"time"

	"github.com/muidea/magicCommon/execute"
)

// Task 任务对象
type Task interface {
	Run()
}

type BackgroundRoutine interface {
	AsyncTask(task Task)
	SyncTask(task Task)
	SyncTaskWithTimeOut(task Task, timeout time.Duration)
	Timer(task Task, intervalValue time.Duration, offsetValue time.Duration)
}

type syncTask struct {
	resultChannel chan bool
	rawTask       Task
	syncTimout    bool
}

func (s *syncTask) Run() {
	s.rawTask.Run()

	if !s.syncTimout {
		s.resultChannel <- true
	}
}

func (s *syncTask) Wait(timeout time.Duration) {
	switch timeout {
	case -1:
		<-s.resultChannel
	default:
		select {
		case <-s.resultChannel:
		case <-time.After(timeout):
		}
	}

	close(s.resultChannel)
	s.syncTimout = true
}

type taskChannel chan Task

// backgroundRoutine backGround routine
type backgroundRoutine struct {
	execute.Execute

	taskChannel taskChannel
}

// NewBackgroundRoutine new Background routine
func NewBackgroundRoutine(capacitySize int) BackgroundRoutine {
	bg := &backgroundRoutine{
		Execute:     execute.NewExecute(capacitySize),
		taskChannel: make(taskChannel),
	}

	bg.run()

	return bg
}

func (s *backgroundRoutine) run() {
	s.Execute.Run(s.loop)
}

func (s *backgroundRoutine) loop() {
	for {
		task, ok := <-s.taskChannel
		if ok {
			s.Execute.Run(func() {
				task.Run()
			})
		}
	}
}

func (s *backgroundRoutine) AsyncTask(task Task) {
	s.Execute.Run(func() {
		s.taskChannel <- task
	})
}

func (s *backgroundRoutine) SyncTask(task Task) {
	s.SyncTaskWithTimeOut(task, -1)
}

func (s *backgroundRoutine) SyncTaskWithTimeOut(task Task, timeout time.Duration) {
	st := &syncTask{rawTask: task, resultChannel: make(chan bool)}
	s.Execute.Run(func() {
		s.taskChannel <- st
	})

	st.Wait(timeout)
}

const onDayDuration = 24 * time.Hour

// Timer exec timer task
func (s *backgroundRoutine) Timer(task Task, intervalValue time.Duration, offsetValue time.Duration) {
	go func() {
		curOffset := func() time.Duration {
			now := time.Now()
			nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
			if intervalValue < 24*time.Hour {
				return (nowOffset/intervalValue+1)*intervalValue - nowOffset
			}

			return (offsetValue + intervalValue - nowOffset + onDayDuration) % onDayDuration
		}()

		//expire := offsetValue + time.Duration(23-now.Hour())*time.Hour + time.Duration(59-now.Minute())*time.Minute + time.Duration(60-now.Second())*time.Second
		time.Sleep(curOffset)

		// 立即执行一次，然后根据周期来执行
		task.Run()

		timeOutTimer := time.NewTicker(intervalValue)
		defer timeOutTimer.Stop()
		for range timeOutTimer.C {
			task.Run()
		}
	}()
}
