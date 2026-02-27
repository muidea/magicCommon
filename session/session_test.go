package session

import (
	"testing"

	"log/slog"
)

func TestUUID(t *testing.T) {
	ids := map[string]string{}

	for idx := 0; idx < 10; idx++ {
		id := createUUID()
		_, ok := ids[id]
		if ok {
			t.Errorf("duplicate id")
			break
		}

		t.Logf("total size:%d, current id:%s", len(ids), id)
		ids[id] = id
	}

	slog.Info("total size:len(ids", "field", len(ids))
}
