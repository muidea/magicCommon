package execute

import (
	"math"

	"github.com/muidea/magicCommon/foundation/log"
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
		log.Warnf("execute queue is full, length:%d, capacity:%d", s.queueLength, s.capacitySize)
	} else if s.queueLength >= int(math.Floor(float64(s.capacitySize)*0.8)) {
		log.Warnf("queue lengths are at warning levels, length:%d, capacity:%d", s.queueLength, s.capacitySize)
	}

	s.capacityQueue <- true
	s.queueLength++
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stackInfo := util.GetStack(3)
				log.Errorf("PANIC: %v\n%s", err, stackInfo)
			}

			<-s.capacityQueue
		}()

		funcPtr()
		s.queueLength--
	}()
}
