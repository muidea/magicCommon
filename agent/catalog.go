package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryCatalog(id int, authToken, sessionID string) (model.CatalogDetailView, bool) {
	result := &common_def.QueryCatalogResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Catalog, true
	}

	log.Printf("query catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) CreateCatalog(name, description string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.CreateCatalogParam{Name: name, Description: description, Catalog: catalog}
	result := &common_def.CreateCatalogResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/catalog/", authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Catalog, true
	}

	log.Printf("create catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) UpdateCatalog(id int, name, description string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateCatalogParam{Name: name, Description: description, Catalog: catalog}
	result := &common_def.UpdateCatalogResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update catalog failed, err:%s", err.Error())
		return result.Catalog, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Catalog, true
	}

	log.Printf("update catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Catalog, false
}

func (s *center) DeleteCatalog(id int, authToken, sessionID string) bool {
	result := &common_def.DestroyCatalogResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/catalog", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete catalog failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	log.Printf("delete catalog failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
