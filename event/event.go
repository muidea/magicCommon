package event

type baseEvent struct {
	eventID          string
	eventSource      string
	eventDestination string
	eventData        interface{}
}

func NewEvent(id, source, destination string, data interface{}) Event {
	return &baseEvent{
		eventID:          id,
		eventSource:      source,
		eventDestination: destination,
		eventData:        data,
	}
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

func (s *baseEvent) Data() interface{} {
	return s.eventData
}
