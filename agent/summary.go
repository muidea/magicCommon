package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) FetchSummary(name, summaryType, authToken, sessionID string) (model.SummaryView, bool) {
	type fetchResult struct {
		common_result.Result
		Summary model.SummaryView `json:"summary"`
	}

	result := &fetchResult{}
	url := fmt.Sprintf("%s/%s?name=%s&type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/", name, summaryType, authToken, sessionID)
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

func (s *center) QuerySummaryContent(id int, summaryType, authToken, sessionID string) []model.SummaryView {
	type queryResult struct {
		common_result.Result
		Summary []model.SummaryView `json:"summary"`
	}

	result := &queryResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/detail", id, summaryType, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}

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

func (s *center) QuerySummaryContentByCatalog(id int, summaryType string, catalog int, authToken, sessionID string) []model.SummaryView {
	type queryResult struct {
		common_result.Result
		Summary []model.SummaryView `json:"summary"`
	}

	result := &queryResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&catalog=%d&authToken=%s&sessionID=%s", s.baseURL, "content/summary/detail", id, summaryType, catalog, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}

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