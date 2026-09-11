package task

import (
	"context"
	"fmt"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/execute"
)

// Task 任务对象
type Task interface {
	Run()
}

type routineTask struct {
	funcPtr func()
}

func (s *routineTask) Run() {
	s.funcPtr()
}

type BackgroundRoutine interface {
	AsyncTask(task Task) error
	// AsyncTaskContext bounds admission only. Accepted tasks must handle their
	// own cancellation and cleanup, even if the context expires while queued.
	AsyncTaskContext(ctx context.Context, task Task) error
	SyncTask(task Task) error
	// SyncTaskWithTimeOut bounds the completion wait after admission. -1 waits
	// indefinitely. A timeout does not cancel an accepted task.
	SyncTaskWithTimeOut(task Task, timeout time.Duration) error
	AsyncFunction(function func()) error
	SyncFunction(function func()) error
	SyncFunctionWithTimeOut(function func(), timeout time.Duration) error
	Timer(ctx context.Context, task Task, intervalValue time.Duration, offsetValue time.Duration) error
	Shutdown(ctx context.Context) bool
}

type syncTask struct {
	resultChannel chan error
	rawTask       Task
}

func (s *syncTask) Run() {
	var err error
	defer func() {
		if value := recover(); value != nil {
			err = cd.NewError(cd.Unexpected, fmt.Sprintf("background task panicked: %v", value))
		}
		// One buffered completion, including panic, even if the waiter has left.
		s.resultChannel <- err
	}()
	s.rawTask.Run()
}

func (s *syncTask) Wait(timeout time.Duration) error {
	if timeout == -1 {
		return <-s.resultChannel
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case err := <-s.resultChannel:
		return err
	case <-timer.C:
		// Prefer an already published completion over a simultaneously fired timer.
		select {
		case err := <-s.resultChannel:
			return err
		default:
			return cd.NewError(cd.Timeout, "background task completion wait timed out; task was not canceled")
		}
	}
}

type taskChannel chan Task

// backgroundRoutine backGround routine
type backgroundRoutine struct {
	execute.Execute

	taskChannel taskChannel
	submitMu    sync.RWMutex
	closed      bool
	closeOnce   sync.Once
	loopDone    chan struct{}
	stopping    chan struct{}
	timers      sync.WaitGroup
	timersDone  chan struct{}
}

// NewBackgroundRoutine new Background routine
func NewBackgroundRoutine(capacitySize int) BackgroundRoutine {
	if capacitySize <= 0 {
		capacitySize = 10
	}
	bg := &backgroundRoutine{
		Execute:     execute.NewExecute(capacitySize),
		taskChannel: make(taskChannel, capacitySize),
		loopDone:    make(chan struct{}),
		stopping:    make(chan struct{}),
		timersDone:  make(chan struct{}),
	}

	bg.run()

	return bg
}

func (s *backgroundRoutine) run() {
	// The dispatcher is tracked by loopDone, not the worker pool: otherwise
	// it consumes the only worker at capacity one and cannot dispatch any job.
	go s.loop()
}

func (s *backgroundRoutine) loop() {
	defer close(s.loopDone)
	for task := range s.taskChannel {
		s.Run(func() {
			task.Run()
		})
	}
}

func (s *backgroundRoutine) AsyncTask(task Task) error {
	return s.submitTask(task)
}

func (s *backgroundRoutine) AsyncTaskContext(ctx context.Context, task Task) error {
	if ctx == nil || ctx.Err() != nil {
		return fmt.Errorf("background task admission context is unavailable")
	}
	return s.submitTaskContext(ctx, task)
}

func (s *backgroundRoutine) SyncTask(task Task) error {
	return s.SyncTaskWithTimeOut(task, -1)
}

func (s *backgroundRoutine) SyncTaskWithTimeOut(task Task, timeout time.Duration) error {
	if task == nil || (timeout < 0 && timeout != -1) {
		return cd.NewError(cd.IllegalParam, "task is required and timeout must be non-negative or -1")
	}
	st := &syncTask{rawTask: task, resultChannel: make(chan error, 1)}
	if err := s.submitTask(st); err != nil {
		return err
	}

	return st.Wait(timeout)
}

func (s *backgroundRoutine) AsyncFunction(function func()) error {
	if function == nil {
		return fmt.Errorf("function is nil")
	}
	return s.AsyncTask(&routineTask{funcPtr: function})
}

func (s *backgroundRoutine) SyncFunction(function func()) error {
	if function == nil {
		return fmt.Errorf("function is nil")
	}
	return s.SyncTask(&routineTask{funcPtr: function})
}

func (s *backgroundRoutine) SyncFunctionWithTimeOut(function func(), timeout time.Duration) error {
	if function == nil {
		return fmt.Errorf("function is nil")
	}
	return s.SyncTaskWithTimeOut(&routineTask{funcPtr: function}, timeout)
}

const onDayDuration = 24 * time.Hour

func (s *backgroundRoutine) Timer(ctx context.Context, task Task, intervalValue time.Duration, offsetValue time.Duration) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	if task == nil {
		return fmt.Errorf("task is nil")
	}
	if intervalValue <= 0 {
		return fmt.Errorf("intervalValue must be positive")
	}
	s.submitMu.RLock()
	defer s.submitMu.RUnlock()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	select {
	case <-s.stopping:
		return fmt.Errorf("background routine is closed")
	default:
	}
	if s.closed {
		return fmt.Errorf("background routine is closed")
	}
	s.timers.Add(1)

	go func() {
		defer s.timers.Done()
		curOffset := func() time.Duration {
			now := time.Now()
			nowOffset := time.Duration(now.Hour())*time.Hour + time.Duration(now.Minute())*time.Minute + time.Duration(now.Second())*time.Second
			if intervalValue < 24*time.Hour {
				return (nowOffset/intervalValue+1)*intervalValue - nowOffset
			}

			return (offsetValue + intervalValue - nowOffset + onDayDuration) % onDayDuration
		}()

		timer := time.NewTimer(curOffset)
		defer timer.Stop()

		select {
		case <-s.stopping:
			return
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		if err := s.AsyncTaskContext(ctx, task); err != nil {
			return
		}

		timeOutTimer := time.NewTicker(intervalValue)
		defer timeOutTimer.Stop()
		for {
			select {
			case <-s.stopping:
				return
			case <-ctx.Done():
				return
			case <-timeOutTimer.C:
				if err := s.AsyncTaskContext(ctx, task); err != nil {
					return
				}
			}
		}
	}()

	return nil
}

func (s *backgroundRoutine) Shutdown(ctx context.Context) bool {
	if ctx == nil {
		ctx = context.Background()
	}
	s.closeOnce.Do(func() {
		// Wake blocked submitters before taking the exclusive close lock.
		close(s.stopping)
		s.submitMu.Lock()
		s.closed = true
		close(s.taskChannel)
		s.submitMu.Unlock()
		go func() { s.timers.Wait(); close(s.timersDone) }()
	})

	select {
	case <-s.timersDone:
	case <-ctx.Done():
		return false
	}
	select {
	case <-s.loopDone:
	case <-ctx.Done():
		return false
	}
	return s.WaitContext(ctx)
}

func (s *backgroundRoutine) submitTask(task Task) error {
	return s.submitTaskContext(context.Background(), task)
}

func (s *backgroundRoutine) submitTaskContext(ctx context.Context, task Task) error {
	if task == nil {
		return fmt.Errorf("task is nil")
	}

	s.submitMu.RLock()
	defer s.submitMu.RUnlock()

	if s.closed {
		return fmt.Errorf("background routine is closed")
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.stopping:
		return fmt.Errorf("background routine is closed")
	case s.taskChannel <- task:
		return nil
	}
}
