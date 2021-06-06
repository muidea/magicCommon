package event

const (
	waitExecute = iota
	finishExecute
)

type baseEvent struct {
	eventID          string
	eventSource      string
	eventDestination string
	eventHeader      Values
	eventData        interface{}
	eventResult      interface{}
}

type baseResult struct {
	statusCode int
	resultData interface{}
}

func NewEvent(id, source, destination string, header Values, data interface{}) Event {
	return &baseEvent{
		eventID:          id,
		eventSource:      source,
		eventDestination: destination,
		eventHeader:      header,
		eventData:        data,
	}
}

func NewResult() Result {
	return &baseResult{statusCode: waitExecute}
}

func (s *baseEvent) ID() string {
	return s.eventID
}

func (s *baseEvent) Source() string {
	return s.eventSource
}

func (s *baseEvent) Destination() string {
	return s.eventDestination
}

func (s *baseEvent) Header() Values {
	return s.eventHeader
}

func (s *baseEvent) Data() interface{} {
	return s.eventData
}

func (s *baseEvent) Result(result interface{}) {
	s.eventResult = result
}

func (s *baseEvent) Match(pattern string) bool {
	return matchID(pattern, s.eventID)
}

func (s *baseResult) Set(data interface{}) {
	s.statusCode = finishExecute
	s.resultData = data
}

func (s *baseResult) IsOK() bool {
	return s.statusCode == finishExecute
}

func (s *baseResult) Get() interface{} {
	return s.resultData
}
