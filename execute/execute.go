package execute

import (
	"log/slog"
	"math"

	"github.com/muidea/magicCommon/foundation/util"
)

type Execute struct {
	capacityQueue chan bool
	queueLength   int
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
	if s.queueLength >= s.capacitySize {
		slog.Warn("execute queue is full, length:s.queueLength, capacity:s.capacitySize", "field", s.queueLength, "error", s.capacitySize)
	} else if s.queueLength >= int(math.Floor(float64(s.capacitySize)*0.8)) {
		slog.Warn("queue lengths are at warning levels, length:s.queueLength, capacity:s.capacitySize", "field", s.queueLength, "error", s.capacitySize)
	}

	s.capacityQueue <- true
	s.queueLength++
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stackInfo := util.GetStack(3)
				slog.Error("PANIC: err\nstackInfo", "field", err, "error", stackInfo)
			}

			<-s.capacityQueue
		}()

		funcPtr()
		s.queueLength--
	}()
}
