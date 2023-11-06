package event

import (
	"github.com/muidea/magicCommon/foundation/log"
	"testing"
)

func TestMatchID(t *testing.T) {
	pattern := "/123"
	id := "/12"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/12"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/#"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111/1212"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/111"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/1212/111"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/bcd/1212/111"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/:id/1212/#"
	id = "/123/122/1212/111/2435/765756f/fsd"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/+"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/#"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/abc/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/cde/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/123/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd"
	id = "abc/123/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/abc/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/create/"
	id = "/warehouse/shelf/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/+/create/"
	id = "/warehouse/shelf/create/"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "/#/+/+/create/"
	id = "/warehouse/shelf/create/"
	if MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/notify/+"
	id = "/bill/notify/123"
	if !MatchValue(pattern, id) {
		t.Errorf("MatchValue failed, pattern:%s, id:%s", pattern, id)
	}
}

type eventHandler struct {
	handlerID string
}

func (s *eventHandler) ID() string {
	return s.handlerID
}

func (s *eventHandler) Notify(ev Event, re Result) {
	log.Infof("eventHandler:%v", s.handlerID)
	if re != nil {
		re.Set(ev.Data(), nil)
	}
}

func TestEventHub(t *testing.T) {
	hub := NewHub()

	eventID := "/e001"
	handler001 := &eventHandler{handlerID: "/h001"}
	hub.Subscribe(eventID, handler001)

	handler002 := &eventHandler{handlerID: "/h002"}
	hub.Subscribe(eventID, handler002)

	val := "e001"
	ev := NewEvent(eventID, "/", handler001.ID(), nil, val)
	re := hub.Send(ev)
	dVal, dErr := re.Get()
	if dErr != nil {
		t.Errorf("send event to hub failed, error:%s", dErr.Error())
		return
	}

	if dVal.(string) != val {
		t.Errorf("send event to hub failed")
		return
	}
}
