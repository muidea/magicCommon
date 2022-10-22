package session

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	log "github.com/cihub/seelog"

	"github.com/muidea/magicCommon/foundation/util"
)

type Endpoint struct {
	Endpoint   string `json:"endpoint"`
	IdentifyID string `json:"identifyID"`
	AuthToken  string `json:"authToken"`
}

func signature(endpoint *Endpoint, ctxVal url.Values) Token {
	credentialVal := fmt.Sprintf("Credential=%s/%s/%s", endpoint.Endpoint, endpoint.IdentifyID, time.Now().UTC().Format("2006010215"))

	headers := []string{}
	for k, _ := range ctxVal {
		headers = append(headers, k)
	}
	sort.Sort(sort.StringSlice(headers))
	signedHeadersVal := fmt.Sprintf("SignedHeaders=%s", strings.Join(headers, ";"))

	encryptVal, encryptErr := util.EncryptByAes(strings.Join([]string{credentialVal, signedHeadersVal}, ","), endpoint.AuthToken)
	if encryptErr != nil {
		log.Errorf("EncryptByAes failed, err:%s", encryptErr.Error())
		return ""
	}
	signatureVal := fmt.Sprintf("Signature=%s", encryptVal)

	return Token(strings.Join([]string{credentialVal, signedHeadersVal, signatureVal}, ","))
}

func decodeCredential(val string) (endpoint string, identifyID string, err error) {
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

	endpoint = items[0]
	identifyID = items[1]
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
