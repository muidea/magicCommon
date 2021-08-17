package event

import "testing"

func TestMatchID(t *testing.T) {
	pattern := "/123"
	id := "/12"

	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "/123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "123"
	id = "12"

	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/+"
	id = "12"

	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123"
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

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123"
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

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "/12"

	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = ""
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "/#"
	id = "#"

	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "#"
	id = "123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "#"
	id = "/123"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "123/+/1212#"
	id = "123"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123/122"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	id = "123/122/"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	id = "123/122/1212"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	pattern = "123/+/1212/#"
	id = "123/122/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	id = "123/122/1212/111/1212"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "123/+/+/1212/#"
	id = "123/122/1212/111"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	id = "123/122/1212/1212"
	if matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
	id = "123/122/1212/1212/uu"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "123/#/+/1212/#"
	id = "123/122/1212/111"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}

	pattern = "123/:id/1212/#"
	id = "123/122/1212/111/2435/765756f/fsd"
	if !matchID(pattern, id) {
		t.Errorf("matchID failed, pattern:%s, id:%s", pattern, id)
		return
	}
}
