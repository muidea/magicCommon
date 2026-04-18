package session

import (
	"encoding/json"
	"fmt"

	"github.com/muidea/magicCommon/foundation/util"
)

type Endpoint struct {
	Endpoint string         `json:"endpoint"`
	Context  map[string]any `json:"context"`
}

func EncryptEndpoint(endpoint *Endpoint) (string, error) {
	endpoint.Context[innerSessionID] = createUUID()
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

func decodeEndpointTokenValue(val string) (ret *Endpoint, err error) {
	secretVal := getSecret()
	strVal, strErr := util.DecryptByAes(val, secretVal)
	if strErr != nil {
		err = strErr
		return
	}

	offset := -1
	for idx, ch := range strVal {
		if ch == '/' {
			offset = idx
			break
		}
	}
	if offset == -1 {
		err = fmt.Errorf("illegal endpoint token value")
		return
	}

	ctx := map[string]any{}
	err = json.Unmarshal([]byte(strVal[offset+1:]), &ctx)
	if err != nil {
		err = fmt.Errorf("illegal endpoint token value")
		return
	}

	ret = &Endpoint{
		Endpoint: strVal[:offset],
		Context:  ctx,
	}
	return
}

func decodeEndpointToken(val string) *sessionImpl {
	endpointPtr, err := decodeEndpointTokenValue(val)
	if err != nil || endpointPtr == nil {
		return nil
	}

	sessionPtr := &sessionImpl{context: map[string]any{}, observer: map[string]Observer{}}
	sessionPtr.context[InnerAuthType] = AuthEndpointSession
	for k, v := range endpointPtr.Context {
		if k == innerSessionID {
			if idVal, ok := v.(string); ok {
				sessionPtr.id = idVal
			}
			continue
		}
		sessionPtr.context[k] = v
	}

	return sessionPtr
}
