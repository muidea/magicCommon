package sms

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
)

// Agent SMS访问代理
type Agent interface {
	PostIdentifyCode(receiver, identifyCode string) error
}

// New 新建Agent
func New(smsSvr string) Agent {
	return &sms{httpClient: &http.Client{}, baseURL: fmt.Sprintf("http://%s", smsSvr)}
}

type sms struct {
	httpClient *http.Client
	baseURL    string
}

func (s *sms) PostIdentifyCode(receiver, identifyCode string) error {
	result := &common_def.Result{}
	url := fmt.Sprintf("%s/%s?receiver=%s&identifyCode=%s", s.baseURL, "IdentifyCode", receiver, identifyCode)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("get status failed, err:%s", err.Error())
		return err
	}
	if result.Success() {
		return nil
	}

	return errors.New(result.Reason)
}
