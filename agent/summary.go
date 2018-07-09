package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) FetchSummary(name string) (model.SummaryView, bool) {
	type fetchResult struct {
		common_result.Result
		Summary model.SummaryView `json:"summary"`
	}

	result := &fetchResult{}
	url := fmt.Sprintf("%s/%s?name=%s&authToken=%s&sessionID=%s", s.baseURL, "content/catalog/", name, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("fetch catalog failed, err:%s", err.Error())
		return result.Summary, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Summary, true
	}

	log.Printf("fetch summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary, false
}

func (s *center) QuerySummaryDetail(id int) []model.SummaryView {
	type queryResult struct {
		common_result.Result
		Summary []model.SummaryView `json:"summary"`
	}

	result := &queryResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/summary", id, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_result.Success {
		return result.Summary
	}

	log.Printf("query summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary
}
