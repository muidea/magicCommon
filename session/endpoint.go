package session

import (
	"fmt"
	log "github.com/cihub/seelog"
	"github.com/muidea/magicCommon/foundation/util"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Endpoint struct {
	Endpoint   string `json:"endpoint"`
	IdentifyID string `json:"identifyID"`
	AuthToken  string `json:"authToken"`
}

func signature(endpoint *Endpoint, ctxVal url.Values) Token {
	credentialVal := fmt.Sprintf("Credential=%s/%s/%s", endpoint.Endpoint, endpoint.IdentifyID, time.Now().UTC().Format("200601021504"))

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
