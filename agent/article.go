package agent

import (
	"fmt"
	"log"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/model"
)

func (s *center) QueryArticle(id int, authToken, sessionID string) (model.ArticleDetailView, bool) {
	result := &common_def.QueryArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	return result.Article, false
}

func (s *center) CreateArticle(title, content string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.CreateArticleParam{Title: title, Content: content, Catalog: catalog}
	result := &common_def.CreateArticleResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/article/", authToken, sessionID)

	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	return result.Article, false
}

func (s *center) UpdateArticle(id int, title, content string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateArticleParam{Title: title, Content: content, Catalog: catalog}
	result := &common_def.UpdateArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)

	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update article failed, err:%s", err.Error())
		return result.Article, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Article, true
	}

	return result.Article, false
}

func (s *center) DeleteArticle(id int, authToken, sessionID string) bool {
	result := &common_def.DestoryArticleResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/article", id, authToken, sessionID)

	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete article failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}
