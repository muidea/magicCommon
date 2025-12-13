package session

import (
	"fmt"
	"net/http"
	"os"
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
	register := DefaultRegistry()
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

func TestDecodeEndpointWithoutSignature(t *testing.T) {
	os.Setenv("HMAC_SECRET", "e3bcbe908a384d9ba8e7ac7028a21f75")

	endpointPtr, endpointErr := decodeSignature("Signature=ZnFdJ5CDfo3ICa55HhsDUkS6ybowwcg6x0PBnAIWqqeozooW/tyqT+utejzvHcuxl6vOPq8qezhbaPw9NoRb2d3n6jsf7gQ+xuKV3ZUNGswHrI4mWaFImxhDx1MMK0okueA4gf0cvvOgGPZljFoANbaPu7cvTUa3Ezi6p8S2M8w3jCnOsouWlE1cbY0YF9Lb4cyIc5Jx69aCG8+Mc2Egr9gtgyRoXBMTBsWVVXPweC24u0sWgPwogaVyNb2sxwLtOoaRR3wbl8zEPTnnExDrbZcJs3bdBdc72CdbyeIV6es=")
	assert.Equal(t, nil, endpointErr)
	assert.Equal(t, "defaultEndpoint", endpointPtr.Endpoint)
}
