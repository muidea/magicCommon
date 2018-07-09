package agent

import (
	"fmt"
	"log"
	"net/http"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

// Agent Center访问代理
type Agent interface {
	Start(centerServer, endpointID, authToken string) bool
	Stop()

	LoginAccount(account, password string) (model.AccountOnlineView, string, string, bool)
	LogoutAccount(authToken, sessionID string) bool
	StatusAccount(authToken, sessionID string) (model.AccountOnlineView, bool)

	QuerySummary(catalogID int) []model.SummaryView

	FetchCatalog(name string) (model.CatalogDetailView, bool)
	QueryCatalog(catalogID int) (model.CatalogDetailView, bool)
	CreateCatalog(name, description string, parent []model.Catalog, authToken, sessionID string) (model.SummaryView, bool)
	UpdateCatalog(id int, name, description string, parent []model.Catalog, authToken, sessionID string) (model.SummaryView, bool)
	DeleteCatalog(id int, authToken, sessionID string) bool

	QueryArticle(id int) (model.ArticleDetailView, bool)
	CreateArticle(title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool)
	UpdateArticle(id int, title, content string, catalog []model.Catalog, authToken, sessionID string) (model.SummaryView, bool)
	DeleteArticle(id int, authToken, sessionID string) bool

	QueryLink(id int) (model.LinkDetailView, bool)

	QueryMedia(id int) (model.MediaDetailView, bool)
}

// NewCenterAgent 新建Agent
func NewCenterAgent() Agent {
	return &center{}
}

type center struct {
	httpClient *http.Client
	baseURL    string
	endpointID string
	authToken  string
	sessionID  string
}

func (s *center) Start(centerServer, endpointID, authToken string) bool {
	s.httpClient = &http.Client{}
	s.baseURL = fmt.Sprintf("http://%s", centerServer)
	s.endpointID = endpointID
	s.authToken = authToken

	sessionID, ok := s.verify()
	if !ok {
		return false
	}

	s.sessionID = sessionID
	log.Print("start centerAgent ok")
	return true
}

func (s *center) Stop() {

}

func (s *center) verify() (string, bool) {
	type verifyResult struct {
		common_result.Result
		SessionID string `json:"sessionID"`
	}

	result := &verifyResult{}
	url := fmt.Sprintf("%s/%s/%s?authToken=%s", s.baseURL, "authority/endpoint/verify", s.endpointID, s.authToken)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("verify endpoint failed, err:%s", err.Error())
		return "", false
	}

	if result.ErrorCode == common_result.Success {
		return result.SessionID, true
	}

	log.Printf("verify endpoint failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return "", false
}

func (s *center) QuerySummary(catalogID int) []model.SummaryView {
	type queryResult struct {
		common_result.Result
		Summary []model.SummaryView `json:"summary"`
	}

	result := &queryResult{Summary: []model.SummaryView{}}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/summary", catalogID, s.authToken, s.sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query summary failed, err:%s", err.Error())
		return result.Summary
	}

	if result.ErrorCode == common_result.Success {
		return result.Summary
	}

	log.Printf("query summary failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.Summary
}
