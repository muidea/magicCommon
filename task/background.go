package task

import "time"

// Task 任务对象
type Task interface {
	Run()
}

type SyncTask struct {
	resultChannel chan bool
	rawTask       Task
}

func (s *SyncTask) Run() {
	s.rawTask.Run()

	s.resultChannel <- true
}

func (s *SyncTask) Wait() {
	<-s.resultChannel

	close(s.resultChannel)
}

type taskChannel chan Task

// BackgroundRoutine backGround routine
type BackgroundRoutine struct {
	taskChannel taskChannel
}

// NewBackgroundRoutine new Background routine
func NewBackgroundRoutine() *BackgroundRoutine {
	bg := &BackgroundRoutine{taskChannel: make(taskChannel)}
	bg.run()

	return bg
}

func (s *BackgroundRoutine) run() {
	go s.loop()
}

func (s *BackgroundRoutine) loop() {
	for {
		task := <-s.taskChannel
		task.Run()
	}
}

// Post exec task
func (s *BackgroundRoutine) Post(task Task) {
	s.taskChannel <- task
}

func (s *BackgroundRoutine) Invoke(task Task) {
	syncTask := &SyncTask{rawTask: task, resultChannel: make(chan bool)}
	s.taskChannel <- syncTask

	syncTask.Wait()
}

const onDayDuration = 24 * time.Hour

// Timer exec timer task
func (s *BackgroundRoutine) Timer(task Task, intervalValue time.Duration, offsetValue time.Duration) {
	go func() {
		now := time.Now()
		curOffset := (onDayDuration/intervalValue+1)*intervalValue + offsetValue
		curOffset -= time.Duration(now.Hour()) * time.Hour
		curOffset -= time.Duration(now.Minute()) * time.Minute
		curOffset -= time.Duration(now.Second()) * time.Second

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
