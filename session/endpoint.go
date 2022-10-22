package session

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

type Endpoint struct {
	Endpoint   string `json:"endpoint"`
	IdentifyID string `json:"identifyID"`
	AuthToken  string `json:"authToken"`
}

func signature(endpoint *Endpoint, ctxVal url.Values) Token {
	credentialVal := fmt.Sprintf("Credential=%s/%s/%s", endpoint.Endpoint, endpoint.IdentifyID, endpoint.AuthToken)

	headers := []string{}
	for k, _ := range ctxVal {
		headers = append(headers, k)
	}
	sort.Sort(sort.StringSlice(headers))
	signedHeadersVal := fmt.Sprintf("SignedHeaders=%s", strings.Join(headers, ";"))

	return Token(strings.Join([]string{credentialVal, signedHeadersVal}, ","))
}

func decodeCredential(val string) (endpoint *Endpoint, err error) {
	items := strings.Split(val, "=")
	if len(items) != 2 {
		err = fmt.Errorf("illegal Credential")
		return
	}
	items = strings.Split(items[1], "/")
	if len(items) != 3 {
		err = fmt.Errorf("illegal Credential value")
		return
	}

	endpoint = &Endpoint{
		Endpoint:   items[0],
		IdentifyID: items[1],
		AuthToken:  items[2],
	}
	return
}

func decodeSignedHeaders(val string) (headers []string, err error) {
	items := strings.Split(val, "=")
	if len(items) != 2 {
		err = fmt.Errorf("illegal SignedHeaders")
		return
	}
	headers = strings.Split(items[1], ";")
	return
}

func decodeSignature(val string) (signature string, err error) {
	items := strings.Split(val, "=")
	if len(items) != 2 {
		err = fmt.Errorf("illegal Signature")
		return
	}
	signature = items[1]
	return
}
