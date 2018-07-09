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

func (s *center) CreateLink(name, description, url, logo string, catalog []model.Catalog, creater int) (model.SummaryView, bool) {
	type createParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		URL         string          `json:"url"`
		Logo        string          `json:"logo"`
		Catalog     []model.Catalog `json:"catalog"`
		Creater     int             `json:"creater"`
	}

	type createResult struct {
		common_result.Result
		Link model.SummaryView `json:"link"`
	}

	param := &createParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog, Creater: creater}
	result := &createResult{}
	httpURL := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/link/", s.authToken, s.sessionID)
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

func (s *center) UpdateLink(id int, name, description, url, logo string, catalog []model.Catalog, updater int) (model.SummaryView, bool) {
	type updateParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		URL         string          `json:"url"`
		Logo        string          `json:"logo"`
		Catalog     []model.Catalog `json:"catalog"`
		Updater     int             `json:"updater"`
	}

	type updateResult struct {
		common_result.Result
		Link model.SummaryView `json:"link"`
	}

	param := &updateParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog, Updater: updater}
	result := &updateResult{}
	httpURL := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, s.authToken, s.sessionID)
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

func (s *center) DeleteLink(id int) bool {
	type deleteResult struct {
		common_result.Result
	}

	result := &deleteResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, s.authToken, s.sessionID)
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
