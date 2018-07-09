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
	BindAccount(user *model.User)
	UnbindAccount()

	FetchSummary(name string) (model.SummaryView, bool)
	QuerySummaryDetail(id int) []model.SummaryView

	QueryCatalog(catalogID int) (model.CatalogDetailView, bool)
	CreateCatalog(name, description string, parent []model.Catalog, creater int) (model.SummaryView, bool)
	UpdateCatalog(id int, name, description string, parent []model.Catalog, updater int) (model.SummaryView, bool)
	DeleteCatalog(id int) bool

	QueryArticle(id int) (model.ArticleDetailView, bool)
	CreateArticle(title, content string, catalog []model.Catalog, creater int) (model.SummaryView, bool)
	UpdateArticle(id int, title, content string, catalog []model.Catalog, updater int) (model.SummaryView, bool)
	DeleteArticle(id int) bool

	QueryLink(id int) (model.LinkDetailView, bool)
	CreateLink(name, description, url, logo string, catalog []model.Catalog, creater int) (model.SummaryView, bool)
	UpdateLink(id int, name, description, url, logo string, catalog []model.Catalog, updater int) (model.SummaryView, bool)
	DeleteLink(id int) bool

	QueryMedia(id int) (model.MediaDetailView, bool)
	CreateMedia(name, description, fileToken string, expiration int, catalog []model.Catalog, creater int) (model.SummaryView, bool)
	BatchCreateMedia(media []model.MediaItem, description string, catalog []model.Catalog, expiration, privacy, creater int) ([]model.SummaryView, bool)
	UpdateMedia(id int, name, description, fileToken string, expiration int, catalog []model.Catalog, updater int) (model.SummaryView, bool)
	DeleteMedia(id int) bool
}

// New 新建Agent
func New() Agent {
	return &center{}
}

type center struct {
	httpClient *http.Client
	baseURL    string
	endpointID string
	authToken  string
	sessionID  string
	bindUser   *model.User
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
