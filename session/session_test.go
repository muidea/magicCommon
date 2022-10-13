package session

import (
	"log"
	"testing"
)

type EntityView struct {
	ID    int    `json:"id"`
	EName string `json:"name"`
	EID   int    `json:"eID"`
	EType string `json:"eType"`
}

type RoleLite struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

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

	impl.context[AuthEntity] = &EntityView{
		ID:    123,
		EName: "testAdmin",
		EID:   1,
		EType: "account",
	}

	impl.context[AuthNamespace] = "xijian"
	impl.context[AuthRole] = &RoleLite{
		ID:   123,
		Name: AuthRole,
	}

	sigVal, sigErr := impl.SignedString()
	if sigErr != nil {
		t.Errorf("SignedString failed, err:%s", sigErr.Error())
		return
	}

	log.Printf("%s", sigVal)

	regImpl := &sessionRegistryImpl{}
	session := regImpl.decodeJWT(sigVal)
	if session == nil {
		t.Errorf("decodeSessionImpl failed")
		return
	}

	entityVal, entityOK := session.GetOption(AuthEntity)
	if !entityOK {
		t.Errorf("decodeSessionImpl authEntity failed")
		return
	}

	entityPtr, entityOK := entityVal.(*EntityView)
	if !entityOK {
		t.Errorf("illegal authEntity auth value failed")
		return
	}

	log.Printf("%d,", entityPtr.ID)
}
