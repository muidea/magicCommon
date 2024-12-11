package initator

import (
	"fmt"
	"testing"

	cd "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/task"
	"github.com/stretchr/testify/assert"
)

type Demo interface {
	HelloWorkd()
}

type Demo2 interface {
	HelloWorkd2()
}

type demo struct {
	id     string
	weight int
}

func (s *demo) ID() string {
	fmt.Printf("id:%s\n", s.id)
	return s.id
}

func (s *demo) Weight() int {
	return s.weight
}

func (s *demo) Setup(endpointName string, eventHub event.Hub, backgroundRoutine task.BackgroundRoutine) {
	fmt.Printf("Setup:%s, endpointName:%s\n", s.id, endpointName)
}

func (s *demo) Run() {
	fmt.Printf("Run:%s\n", s.id)
}

func (s *demo) Teardown() {
	fmt.Printf("Teardown:%s\n", s.id)
}

func (s *demo) HelloWorkd() {
	fmt.Printf("HelloWorkd:%s\n", s.id)
}

func NewDemo123() *demo {
	return &demo{
		id:     "123",
		weight: 123,
	}
}

func NewDemo100() *demo {
	return &demo{
		id:     "100",
		weight: 100,
	}
}

func Test_Initator(t *testing.T) {
	d100 := NewDemo100()
	d123 := NewDemo123()

	Register(d100)
	Register(d123)

	Setup(nil, nil)
	Run()

	var demoPtr Demo
	var demo2Ptr Demo2
	var result *cd.Result
	demo2Ptr, demo2Err := GetEntity("123", demo2Ptr)
	assert.NotEqual(t, demo2Err, result)
	assert.Equal(t, demo2Err.Success(), false)
	assert.Equal(t, demo2Ptr, nil)

	demoPtr, demoErr := GetEntity("100", demoPtr)
	assert.Equal(t, demoErr, result)

	demoPtr.HelloWorkd()

	Teardown()
}
