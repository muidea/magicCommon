package session

import (
	"fmt"
	"testing"
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
	val, err := EncryptEndpoint(ptr)
	if err != nil {
		t.Errorf("encrypt endpoint failed, err:%s", err.Error())
		return
	}

	fmt.Printf("%s\n", val)
}

func TestSignatureEndpoint(t *testing.T) {
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

	sessionPtr := DecodeEndpoint(string(token))
	if sessionPtr == nil {
		t.Errorf("decode endpoint failed, err:%s", err.Error())
		return
	}
}
