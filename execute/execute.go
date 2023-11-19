package execute

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

func (s *Execute) Run(funcPtr func()) {
	s.capacityQueue <- true
	go func() {
		funcPtr()
		<-s.capacityQueue
	}()
}
