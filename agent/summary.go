package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) FetchSummary(name, summaryType, authToken, sessionID string) (model.SummaryView, bool) {
	result := &common_def.QuerySummaryResult{}
	url := fmt.Sprintf("%s/%s?name=%s&type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/", name, summaryType, authToken, sessionID)
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("fetch catalog failed, err:%s", err.Error())
		return result.Summary, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, true
	}

	log.Printf("fetch summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary, false
}

func (s *center) QuerySummaryContent(id int, summaryType, authToken, sessionID string) []model.SummaryView {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/detail", id, summaryType, authToken, sessionID)
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary
	}

	log.Printf("query summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary
}

func (s *center) QuerySummaryContentByUser(id int, summaryType, authToken, sessionID string, user int) []model.SummaryView {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&user=%d&authToken=%s&sessionID=%s", s.baseURL, "content/summary/detail", id, summaryType, user, authToken, sessionID)
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary
	}

	log.Printf("query summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary
}
