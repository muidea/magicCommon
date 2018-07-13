package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryLink(id int, authToken, sessionID string) (model.LinkDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Link model.LinkDetailView `json:"link"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)
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

func (s *center) CreateLink(name, description, url, logo string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		URL         string          `json:"url"`
		Logo        string          `json:"logo"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Link model.SummaryView `json:"link"`
	}

	param := &createParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog}
	result := &createResult{}
	httpURL := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/link/", authToken, sessionID)
	err := net.HTTPPost(s.httpClient, httpURL, param, result)
	if err != nil {
		log.Printf("create link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Link, true
	}

	log.Printf("create link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}

func (s *center) UpdateLink(id int, name, description, url, logo string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type updateParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		URL         string          `json:"url"`
		Logo        string          `json:"logo"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type updateResult struct {
		common_result.Result
		Link model.SummaryView `json:"link"`
	}

	param := &updateParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog}
	result := &updateResult{}
	httpURL := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)
	err := net.HTTPPut(s.httpClient, httpURL, param, result)
	if err != nil {
		log.Printf("update link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Link, true
	}

	log.Printf("update link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}

func (s *center) DeleteLink(id int, authToken, sessionID string) bool {
	type deleteResult struct {
		common_result.Result
	}

	result := &deleteResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete link failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_result.Success {
		return true
	}

	log.Printf("query link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
