package session

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var ptr = &Endpoint{
	Endpoint: "test",
	Context: map[string]interface{}{
		"name":   "test",
		"age":    123,
		"enable": true,
	},
}

func TestEncryptEndpoint(t *testing.T) {
	ptr.Context[AuthExpireTime] = time.Now().Add(time.Hour).UTC().UnixMilli()
	val, err := EncryptEndpoint(ptr)
	if err != nil {
		t.Errorf("encrypt endpoint failed, err:%s", err.Error())
		return
	}

	// "l/87eGjl6aHP8VKN+ogxW/cUgZ1r/iE7+jR0ZslJ5s/jYtz9SRCOcr55UcM2c1pUCJnZrkhzYHK/bdY6qaJoEg=="
	fmt.Printf("%s\n", val)
}

func TestSignatureEndpoint(t *testing.T) {
	ptr.Context[AuthExpireTime] = time.Now().Add(time.Hour).UTC().UnixMilli()
	val, err := EncryptEndpoint(ptr)
	if err != nil {
		t.Errorf("encrypt endpoint failed, err:%s", err.Error())
		return
	}

	token, err := SignatureEndpoint(ptr.Endpoint, val)
	if err != nil {
		t.Errorf("signature endpoint failed, err:%s", err.Error())
		return
	}
	fmt.Printf("%s\n", token)
}

func TestDecodeEndpoint(t *testing.T) {
	register := CreateRegistry()
	defer register.Release()

	ptr.Context[AuthExpireTime] = time.Now().Add(time.Hour).UTC().UnixMilli()
	val, err := EncryptEndpoint(ptr)
	if err != nil {
		t.Errorf("encrypt endpoint failed, err:%s", err.Error())
		return
	}

	token, err := SignatureEndpoint(ptr.Endpoint, val)
	if err != nil {
		t.Errorf("signature endpoint failed, err:%s", err.Error())
		return
	}

	req := &http.Request{Header: http.Header{}}
	var res http.ResponseWriter
	req.Header.Set(Authorization, fmt.Sprintf("%s %s", sigToken, token))

	sessionPtr := register.GetSession(res, req)
	assert.NotEqual(t, nil, sessionPtr)
	nameVal, nameOK := sessionPtr.GetString("name")
	assert.Equal(t, true, nameOK)
	assert.Equal(t, "test", nameVal)
	ageVal, ageOK := sessionPtr.GetInt("age")
	assert.Equal(t, true, ageOK)
	assert.Equal(t, int64(123), ageVal)
	enableVal, enableOK := sessionPtr.GetBool("enable")
	assert.Equal(t, true, enableOK)
	assert.Equal(t, true, enableVal)
}
