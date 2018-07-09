package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryLink(id int) (model.LinkDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Link model.LinkDetailView `json:"link"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Link, true
	}

	log.Printf("query link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}
