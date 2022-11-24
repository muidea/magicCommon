package session

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/muidea/magicCommon/foundation/util"
)

const (
	Credential = "Credential"
	Signature  = "Signature"
)

type Endpoint struct {
	Endpoint string                 `json:"endpoint"`
	Context  map[string]interface{} `json:"context"`
}

func EncryptEndpoint(endpoint *Endpoint) (string, error) {
	valData, valErr := json.Marshal(endpoint.Context)
	if valErr != nil {
		return "", valErr
	}

	valStr := fmt.Sprintf("%s/%s", endpoint.Endpoint, string(valData))
	valStr, valErr = util.EncryptByAes(valStr, getSecret())
	if valErr != nil {
		return "", valErr
	}

	return valStr, nil
}

func SignatureEndpoint(endpoint string, authToken string) (Token, error) {
	credentialVal := fmt.Sprintf("%s=%s", Credential, endpoint)
	signatureVal := fmt.Sprintf("%s=%s", Signature, authToken)
	return Token(strings.Join([]string{credentialVal, signatureVal}, ",")), nil
}

func decodeEndpoint(sigVal string) *sessionImpl {
	offset := strings.Index(sigVal, ",")
	if offset == -1 {
		return nil
	}
	endpointVal, valErr := decodeCredential(sigVal[:offset])
	if valErr != nil {
		return nil
	}

	endpointPtr, ptrErr := decodeSignature(sigVal[offset+1:])
	if ptrErr != nil {
		return nil
	}
	if endpointVal != endpointPtr.Endpoint {
		return nil
	}

	sessionPtr := &sessionImpl{context: map[string]interface{}{}, observer: map[string]Observer{}}
	for k, v := range endpointPtr.Context {
		sessionPtr.context[k] = v
	}
	sessionPtr.id = sigVal[offset+1:]

	return sessionPtr
}

func decodeCredential(val string) (ret string, err error) {
	offset := strings.Index(val, "=")
	if offset == -1 {
		err = fmt.Errorf("illegal Credential")
		return
	}

	if val[:offset] != Credential {
		err = fmt.Errorf("illegal Credential head")
		return
	}

	ret = val[offset+1:]
	return
}

func decodeSignature(val string) (ret *Endpoint, err error) {
	offset := strings.Index(val, "=")
	if offset == -1 {
		err = fmt.Errorf("illegal Signature")
		return
	}

	if val[:offset] != Signature {
		err = fmt.Errorf("illegal Signature head")
		return
	}

	strVal, strErr := util.DecryptByAes(val[offset+1:], getSecret())
	if strErr != nil {
		err = strErr
		return
	}

	offset = strings.Index(strVal, "/")
	if offset == -1 {
		err = fmt.Errorf("illegal Signature value")
		return
	}

	ctx := map[string]interface{}{}
	err = json.Unmarshal([]byte(strVal[offset+1:]), &ctx)
	if err != nil {
		err = fmt.Errorf("illegal Signature value")
		return
	}

	ret = &Endpoint{
		Endpoint: strVal[:offset],
		Context:  ctx,
	}
	return
}
