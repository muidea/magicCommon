package task

import (
	"time"
)

// Task 任务对象
type Task interface {
	Run()
}

type BackgroundRoutine interface {
	Post(task Task)
	Invoke(task Task)
	Timer(task Task, intervalValue time.Duration, offsetValue time.Duration)
}

type syncTask struct {
	resultChannel chan bool
	rawTask       Task
}

func (s *syncTask) Run() {
	s.rawTask.Run()

	s.resultChannel <- true
}

func (s *syncTask) Wait() {
	<-s.resultChannel

	close(s.resultChannel)
}

type taskChannel chan Task

// backgroundRoutine backGround routine
type backgroundRoutine struct {
	taskChannel taskChannel
}

// NewBackgroundRoutine new Background routine
func NewBackgroundRoutine() BackgroundRoutine {
	bg := &backgroundRoutine{taskChannel: make(taskChannel)}
	bg.run()

	return bg
}

func (s *backgroundRoutine) run() {
	go s.loop()
}

func (s *backgroundRoutine) loop() {
	for {
		task := <-s.taskChannel
		task.Run()
	}
}

// Post exec task
func (s *backgroundRoutine) Post(task Task) {
	go func() {
		s.taskChannel <- task
	}()
}

func (s *backgroundRoutine) Invoke(task Task) {
	syncTask := &syncTask{rawTask: task, resultChannel: make(chan bool)}
	go func() {
		s.taskChannel <- syncTask
	}()

	syncTask.Wait()
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

			return (offsetValue + intervalValue - nowOffset + 24*time.Hour) % (24 * time.Hour)
		}()

		//expire := offsetValue + time.Duration(23-now.Hour())*time.Hour + time.Duration(59-now.Minute())*time.Minute + time.Duration(60-now.Second())*time.Second
		time.Sleep(curOffset)

		// 立即执行一次，然后根据周期来执行
		task.Run()

		timeOutTimer := time.NewTicker(intervalValue)
		for {
			select {
			case <-timeOutTimer.C:
				task.Run()
			}
		}
	}()
}
