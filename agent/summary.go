package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/foundation/util"
	"muidea.com/magicCommon/model"
)

func (s *center) FetchSummary(summaryName, summaryType, authToken, sessionID string, strictCatalog *model.CatalogUnit) (model.SummaryView, bool) {
	result := &common_def.QuerySummaryResult{}
	url := fmt.Sprintf("%s/%s?name=%s&type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/", summaryName, summaryType, authToken, sessionID)
	if strictCatalog != nil {
		url = fmt.Sprintf("%s&%s", url, common_def.EncodeStrictCatalog(*strictCatalog))
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("fetch catalog failed, err:%s", err.Error())
		return result.Summary, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, true
	}

	return result.Summary, false
}

func (s *center) QuerySummaryContent(summary model.CatalogUnit, authToken, sessionID string) []model.SummaryView {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary", summary.ID, summary.Type, authToken, sessionID)

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary
	}

	return result.Summary
}

func (s *center) QuerySummaryContentWithCatalog(summary model.CatalogUnit, authToken, sessionID string, strictCatalog *model.CatalogUnit) []model.SummaryView {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary", summary.ID, summary.Type, authToken, sessionID)
	if strictCatalog != nil {
		url = fmt.Sprintf("%s&%s", url, common_def.EncodeStrictCatalog(*strictCatalog))
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary
	}

	return result.Summary
}

func (s *center) QuerySummaryContentByUser(user int, authToken, sessionID string, strictCatalog *model.CatalogUnit) []model.SummaryView {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s?user[]=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summarys/", util.IntArray2Str([]int{user}), authToken, sessionID)
	if strictCatalog != nil {
		url = fmt.Sprintf("%s&%s", url, common_def.EncodeStrictCatalog(*strictCatalog))
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary
	}

	return result.Summary
}
