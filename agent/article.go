package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QueryArticle(id int) (model.ArticleDetailView, bool) {
	type queryResult struct {
		common_result.Result
		Article model.ArticleDetailView `json:"article"`
	}

	result := &queryResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Article, true
	}

	log.Printf("query article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) CreateArticle(title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name    string          `json:"name"`
		Content string          `json:"content"`
		Catalog []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Article model.SummaryView `json:"article"`
	}

	param := &createParam{Name: title, Content: content, Catalog: catalog}
	result := &createResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/article/", authToken, sessionID)
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Article, true
	}

	log.Printf("create article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) UpdateArticle(id int, title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool) {
	type createParam struct {
		Name    string          `json:"name"`
		Content string          `json:"content"`
		Catalog []model.Catalog `json:"catalog"`
	}

	type createResult struct {
		common_result.Result
		Article model.SummaryView `json:"article"`
	}

	param := &createParam{Name: title, Content: content, Catalog: catalog}
	result := &createResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)
	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_result.Success {
		return result.Article, true
	}

	log.Printf("update article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Article, false
}

func (s *center) DeleteArticle(id int, authToken, sessionID string) bool {
	type deleteResult struct {
		common_result.Result
	}

	result := &deleteResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, s.authToken, s.sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete article failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_result.Success {
		return true
	}

	log.Printf("query article failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}
