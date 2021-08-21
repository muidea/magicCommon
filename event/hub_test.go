package event

import "testing"

func TestMatchID(t *testing.T) {
	pattern := "/123"
	id := "/12"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123"
	id = "/123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "/12"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/12"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/#"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/1212/#"
	id = "/123/122/1212/111/1212"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/111"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/+/+/1212/#"
	id = "/123/122/1212/1212/uu/www/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/1212/111"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/#/+/1212/#"
	id = "/123/122/abc/bcd/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/123/:id/1212/#"
	id = "/123/122/1212/111/2435/765756f/fsd"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/+"
	id = "/warehouse/shelf/create/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/update/#"
	id = "/warehouse/shelf/create/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/shelf/create/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/abc/create/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/abc/#"
	id = "/abc/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/abc/bcd/cde/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "/abc/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/abc/#/bcd/"
	id = "abc/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/#/bcd/"
	id = "abc/123/bcd/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd/"
	id = "abc/123/bcd"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/bcd"
	id = "abc/123/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/abc/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "abc/+/+/bcd"
	id = "abc/123/bcd/bcd"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/create/"
	id = "/warehouse/shelf/create/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#/+/create/"
	id = "/warehouse/shelf/create/"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "/#/+/+/create/"
	id = "/warehouse/shelf/create/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
}
