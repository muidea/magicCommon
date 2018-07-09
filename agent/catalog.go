package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) FetchCatalog(name string) (model.CatalogDetailView, bool) {
	type fetchResult struct {
		common_result.Result
		Catalog model.CatalogDetailView `json:"catalog"`
	}

	result := &fetchResult{}
	url := fmt.Sprintf("%s/%s?name=%s&authToken=%s&sessionID=%s", s.baseURL, "content/catalog/", name, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("fetch catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Catalog, true
	}

	log.Printf("fetch catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) QueryCatalog(catalogID int) (model.CatalogDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Catalog model.CatalogDetailView `json:"catalog"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", catalogID, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Catalog, true
	}

	log.Printf("query catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) CreateCatalog(name, description string, parent []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Catalog model.SummaryView `json:"catalog"`
	}

	param := &createParam{Name: name, Description: description, Catalog: parent}
	result := &createResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/catalog/", authToken, sessionID)
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Catalog, true
	}

	log.Printf("create catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) UpdateCatalog(id int, name, description string, parent []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Catalog model.SummaryView `json:"catalog"`
	}

	param := &createParam{Name: name, Description: description, Catalog: parent}
	result := &createResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", id, authToken, sessionID)
	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Catalog, true
	}

	log.Printf("update catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) DeleteCatalog(id int, authToken, sessionID string) bool {
	type deleteResult struct {
		common_result.Result
	}

	result := &deleteResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", id, s.authToken, s.sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete catalog failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_result.Success {
		return true
	}

	log.Printf("delete catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
