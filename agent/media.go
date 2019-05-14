package agent

import (
	"fmt"
	"log"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/model"
)

func (s *center) QueryMedia(id int, authToken, sessionID string) (model.MediaDetailView, bool) {
	result := &common_def.QueryMediaResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Media, true
	}

	return result.Media, false
}

func (s *center) CreateMedia(name, description, fileToken string, expiration int, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.CreateMediaParam{Name: name, Description: description, FileToken: fileToken, Expiration: expiration, Catalog: catalog}
	result := &common_def.CreateMediaResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/media/", authToken, sessionID)

	_, err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Media, true
	}

	return result.Media, false
}

func (s *center) BatchCreateMedia(media []common_def.MediaInfo, description string, catalog []model.CatalogUnit, expiration int, authToken, sessionID string) ([]model.SummaryView, bool) {
	param := &common_def.BatchCreateMediaParam{Medias: media, Description: description, Expiration: expiration, Catalog: catalog}
	result := &common_def.BatchCreateMediaResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/media/batch/", authToken, sessionID)

	_, err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create media failed, err:%s", err.Error())
		return result.Medias, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Medias, true
	}

	return result.Medias, false
}

func (s *center) UpdateMedia(id int, name, description, fileToken string, expiration int, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateMediaParam{Name: name, Description: description, FileToken: fileToken, Expiration: expiration, Catalog: catalog}
	result := &common_def.UpdateMediaResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)

	_, err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update media failed, err:%s", err.Error())
		return result.Media, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Media, true
	}

	return result.Media, false
}

func (s *center) DeleteMedia(id int, authToken, sessionID string) bool {
	result := &common_def.DestroyMediaResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/media", id, authToken, sessionID)

	_, err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete media failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}
