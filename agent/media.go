package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryMedia(id int, authToken, sessionID string) (model.MediaDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Media model.MediaDetailView `json:"media"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Media, true
	}

	log.Printf("query media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Media, false
}

func (s *center) CreateMedia(name, description, fileToken string, expiration, privacy int, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		FileToken   string          `json:"fileToken"`
		Expiration  int             `json:"expiration"`
		Privacy     int             `json:"privacy"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Media model.SummaryView `json:"media"`
	}

	param := &createParam{Name: name, Description: description, FileToken: fileToken, Expiration: expiration, Privacy: privacy, Catalog: catalog}
	result := &createResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/media/", authToken, sessionID)
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Media, true
	}

	log.Printf("create media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Media, false
}

func (s *center) BatchCreateMedia(media []model.MediaItem, description string, catalog []model.Catalog, expiration, privacy int, authToken, sessionID string) ([]model.SummaryView, bool) {
	type batchCreateParam struct {
		Media       []model.MediaItem `json:"media"`
		Description string            `json:"description"`
		Expiration  int               `json:"expiration"`
		Privacy     int               `json:"privacy"`
		Catalog     []model.Catalog   `json:"catalog"`
	}

	type batchCreateResult struct {
		common_result.Result
		Media []model.SummaryView `json:"media"`
	}

	param := &batchCreateParam{Media: media, Description: description, Expiration: expiration, Privacy: privacy, Catalog: catalog}
	result := &batchCreateResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/media/", authToken, sessionID)
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Media, true
	}

	log.Printf("create media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Media, false
}

func (s *center) UpdateMedia(id int, name, description, fileToken string, expiration, privacy int, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type updateParam struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		FileToken   string          `json:"fileToken"`
		Expiration  int             `json:"expiration"`
		Privacy     int             `json:"privacy"`
		Catalog     []model.Catalog `json:"catalog"`
	}

	type updateResult struct {
		common_result.Result
		Media model.SummaryView `json:"media"`
	}

	param := &updateParam{Name: name, Description: description, FileToken: fileToken, Expiration: expiration, Privacy: privacy, Catalog: catalog}
	result := &updateResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)
	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Media, true
	}

	log.Printf("update media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Media, false
}

func (s *center) DeleteMedia(id int, authToken, sessionID string) bool {
	type deleteResult struct {
		common_result.Result
	}

	result := &deleteResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete media failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_result.Success {
		return true
	}

	log.Printf("query media failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
