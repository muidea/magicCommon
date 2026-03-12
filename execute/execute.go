package execute

import (
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

type Execute struct {
	mu            sync.Mutex
	capacityQueue chan bool
	queueLength   int
	activeCount   int
	capacitySize  int
}

func NewExecute(capacitySize int) Execute {
	if capacitySize <= 0 {
		capacitySize = 10
	}

	return Execute{
		capacitySize:  capacitySize,
		capacityQueue: make(chan bool, capacitySize),
	}
}

func (s *Execute) Lock() { /* for noCopy */ }

func (s *Execute) Unlock() { /* for noCopy */ }

func (s *Execute) Run(funcPtr func()) {
	s.mu.Lock()
	if s.queueLength >= s.capacitySize {
		s.mu.Unlock()
		slog.Warn("execute queue is full, length:s.queueLength, capacity:s.capacitySize", "field", s.queueLength, "error", s.capacitySize)
	} else if s.queueLength >= int(math.Floor(float64(s.capacitySize)*0.8)) {
		s.mu.Unlock()
		slog.Warn("queue lengths are at warning levels, length:s.queueLength, capacity:s.capacitySize", "field", s.queueLength, "error", s.capacitySize)
	} else {
		s.mu.Unlock()
	}

	s.capacityQueue <- true
	s.mu.Lock()
	s.queueLength++
	s.activeCount++
	s.mu.Unlock()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stackInfo := util.GetStack(3)
				slog.Error("PANIC: err\nstackInfo", "field", err, "error", stackInfo)
			}

			<-s.capacityQueue
			s.mu.Lock()
			s.queueLength--
			s.activeCount--
			s.mu.Unlock()
		}()

		funcPtr()
	}()
}

func (s *Execute) Wait() {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		if s.activeCount == 0 {
			s.mu.Unlock()
			return
		}
		s.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}
