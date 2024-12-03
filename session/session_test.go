package session

import (
	"testing"

	"github.com/muidea/magicCommon/foundation/log"
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

	log.Infof("total size:%d", len(ids))
}
