package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryLink(id int, authToken, sessionID string) (model.LinkDetailView, bool) {
	result := &common_def.QueryLinkResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Link, true
	}

	log.Printf("query link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}

func (s *center) CreateLink(name, description, url, logo string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.CreateLinkParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog}
	result := &common_def.CreateLinkResult{}
	httpURL := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/link/", authToken, sessionID)

	err := net.HTTPPost(s.httpClient, httpURL, param, result)
	if err != nil {
		log.Printf("create link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Link, true
	}

	log.Printf("create link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}

func (s *center) UpdateLink(id int, name, description, url, logo string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateLinkParam{Name: name, Description: description, URL: url, Logo: logo, Catalog: catalog}
	result := &common_def.UpdateLinkResult{}
	httpURL := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)

	err := net.HTTPPut(s.httpClient, httpURL, param, result)
	if err != nil {
		log.Printf("update link failed, err:%s", err.Error())
		return result.Link, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Link, true
	}

	log.Printf("update link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Link, false
}

func (s *center) DeleteLink(id int, authToken, sessionID string) bool {
	result := &common_def.DestroyLinkResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/link", id, authToken, sessionID)

	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete link failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	log.Printf("query link failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
