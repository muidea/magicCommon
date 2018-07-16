package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryArticle(id int, authToken, sessionID string) (model.ArticleDetailView, bool) {
	result := &common_def.QueryArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	log.Printf("query article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) CreateArticle(title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.CreateArticleParam{Name: title, Content: content, Catalog: catalog}
	result := &common_def.CreateArticleResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/article/", authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	log.Printf("create article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) UpdateArticle(id int, title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateArticleParam{Name: title, Content: content, Catalog: catalog}
	result := &common_def.UpdateArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	log.Printf("update article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) DeleteArticle(id int, authToken, sessionID string) bool {
	result := &common_def.DestoryArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)
	if s.bindUser != nil {
		url = fmt.Sprintf("%s&user=%d", url, s.bindUser.ID)
	}
	if s.strictCatalog != nil {
		url = fmt.Sprintf("%s&strictCatalog=%d", url, s.strictCatalog.ID)
	}

	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete article failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	log.Printf("query article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
