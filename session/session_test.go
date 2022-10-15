package session

import (
	"log"
	"testing"
)

func TestUUID(t *testing.T) {
	ids := map[string]string{}

	for idx := 0; idx < 10000000; idx++ {
		id := createUUID()
		_, ok := ids[id]
		if ok {
			t.Errorf("duplicate id")
			break
		}

		t.Logf("total size:%d, current id:%s", len(ids), id)
		ids[id] = id
	}

	log.Printf("total size:%d", len(ids))
}

func TestSessionImpl_SignedString(t *testing.T) {
	impl := &sessionImpl{
		id:      createUUID(),
		context: map[string]interface{}{},
	}

	impl.context["AuthEntity"] = 123

	impl.context["Namespace"] = "xijian"
	impl.context["AuthRole"] = 123.456

	sigVal := impl.SignedString()

	log.Printf("%s", sigVal)

	regImpl := &sessionRegistryImpl{}
	session := regImpl.decodeJWT(sigVal)
	if session == nil {
		t.Errorf("decodeSessionImpl failed")
		return
	}

	entityVal, entityOK := session.GetInt("AuthEntity")
	if !entityOK {
		t.Errorf("decodeSessionImpl authEntity failed")
		return
	}

	if entityVal != 123 {
		t.Errorf("illegal authEntity auth value failed")
		return
	}

	roleVal, roleOK := session.GetFloat("AuthRole")
	if !roleOK {
		t.Errorf("decodeSessionImpl AuthRole failed")
		return

	}
	if roleVal != 123.456 {
		t.Errorf("decodeSessionImpl AuthRole failed")
		return
	}
}
