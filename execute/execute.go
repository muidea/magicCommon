package execute

import (
	"github.com/muidea/magicCommon/foundation/log"
	"github.com/muidea/magicCommon/foundation/util"
)

type Execute struct {
	capacityQueue chan bool
}

func NewExecute(capacitySize int) Execute {
	if capacitySize <= 0 {
		capacitySize = 10
	}

	return Execute{
		capacityQueue: make(chan bool, capacitySize),
	}
}

func (s *Execute) Lock() { /* for noCopy */ }

func (s *Execute) Unlock() { /* for noCopy */ }

func (s *Execute) Run(funcPtr func()) {
	s.capacityQueue <- true
	go func() {
		defer func() {
			if err := recover(); err != nil {
				stackInfo := util.GetStack(3)
				log.Errorf("PANIC: %v\n%s", err, stackInfo)
			}

			<-s.capacityQueue
		}()

		funcPtr()
	}()
}
