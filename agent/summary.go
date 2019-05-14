package agent

import (
	"fmt"
	"log"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicCommon/model"
)

func (s *center) FetchSummary(summaryName, summaryType, authToken, sessionID string, strictCatalog *model.CatalogUnit) (model.SummaryView, bool) {
	result := &common_def.QuerySummaryResult{}
	url := fmt.Sprintf("%s/%s?name=%s&type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary/", summaryName, summaryType, authToken, sessionID)
	if strictCatalog != nil {
		strictVal := common_def.EncodeStrictCatalog(*strictCatalog)
		if strictVal != "" {
			url = fmt.Sprintf("%s&%s", url, strictVal)
		}
	}

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("fetch catalog failed, err:%s", err.Error())
		return result.Summary, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, true
	}

	return result.Summary, false
}

func (s *center) QuerySummaryContent(summary model.CatalogUnit, filter *common_def.Filter, authToken, sessionID string) ([]model.SummaryView, int) {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary", summary.ID, summary.Type, authToken, sessionID)
	if filter != nil {
		filterVal := filter.Encode()
		if filterVal != "" {
			url = fmt.Sprintf("%s&%s", url, filterVal)
		}
	}

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary, result.Total
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, result.Total
	}

	return result.Summary, result.Total
}

func (s *center) QuerySummaryContentWithSpecialType(summary model.CatalogUnit, specialType string, filter *common_def.Filter, authToken, sessionID string) ([]model.SummaryView, int) {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?type=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summary", summary.ID, summary.Type, authToken, sessionID)
	if specialType != "" {
		url = fmt.Sprintf("%s&specialType=%s", url, specialType)
	}
	if filter != nil {
		filterVal := filter.Encode()
		if filterVal != "" {
			url = fmt.Sprintf("%s&%s", url, filterVal)
		}
	}

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary, result.Total
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, result.Total
	}

	return result.Summary, result.Total
}

func (s *center) QuerySummaryContentByUser(user int, filter *common_def.Filter, authToken, sessionID string, strictCatalog *model.CatalogUnit) ([]model.SummaryView, int) {
	result := &common_def.QuerySummaryListResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s?user[]=%s&authToken=%s&sessionID=%s", s.baseURL, "content/summarys/", util.IntArray2Str([]int{user}), authToken, sessionID)
	if strictCatalog != nil {
		strictVal := common_def.EncodeStrictCatalog(*strictCatalog)
		if strictVal != "" {
			url = fmt.Sprintf("%s&%s", url, common_def.EncodeStrictCatalog(*strictCatalog))
		}
	}
	if filter != nil {
		filterVal := filter.Encode()
		if filterVal != "" {
			url = fmt.Sprintf("%s&%s", url, filterVal)
		}
	}

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary, result.Total
	}

	if result.ErrorCode == common_def.Success {
		return result.Summary, result.Total
	}

	return result.Summary, result.Total
}
