package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryMedia(id int) (model.MediaDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Media model.MediaDetailView `json:"media"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Media, true
	}

	log.Printf("query media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Media, false
}
