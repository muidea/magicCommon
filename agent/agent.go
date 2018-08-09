package agent

import (
	"fmt"
	"log"
	"net/http"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

// Agent Center访问代理
type Agent interface {
	Start(centerServer, endpointID, authToken string) (string, string, bool)
	Stop()

	LoginAccount(account, password string) (model.OnlineEntryView, string, string, bool)
	LogoutAccount(authToken, sessionID string) bool
	StatusAccount(authToken, sessionID string) (model.OnlineEntryView, string, string, bool)

	StrictCatalog(catalog *model.CatalogUnit)
	UnstrictCatalog()

	FetchSummary(name, summaryType, authToken, sessionID string) (model.SummaryView, bool)
	QuerySummaryContent(id int, summaryType, authToken, sessionID string) []model.SummaryView
	QuerySummaryContentByUser(id int, summaryType, authToken, sessionID string, user int) []model.SummaryView

	QueryCatalog(id int, authToken, sessionID string) (model.CatalogDetailView, bool)
	CreateCatalog(name, description string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	UpdateCatalog(id int, name, description string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	DeleteCatalog(id int, authToken, sessionID string) bool

	QueryArticle(id int, authToken, sessionID string) (model.ArticleDetailView, bool)
	CreateArticle(title, content string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	UpdateArticle(id int, title, content string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	DeleteArticle(id int, authToken, sessionID string) bool

	QueryLink(id int, authToken, sessionID string) (model.LinkDetailView, bool)
	CreateLink(name, description, url, logo string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	UpdateLink(id int, name, description, url, logo string, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	DeleteLink(id int, authToken, sessionID string) bool

	QueryMedia(id int, authToken, sessionID string) (model.MediaDetailView, bool)
	CreateMedia(name, description, fileToken string, expiration int, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	BatchCreateMedia(media []common_def.MediaInfo, description string, catalog []model.CatalogUnit, expiration int, authToken, sessionID string) ([]model.SummaryView, bool)
	UpdateMedia(id int, name, description, fileToken string, expiration int, catalog []model.CatalogUnit, authToken, sessionID string) (model.SummaryView, bool)
	DeleteMedia(id int, authToken, sessionID string) bool
}

// New 新建Agent
func New() Agent {
	return &center{}
}

type center struct {
	httpClient    *http.Client
	baseURL       string
	strictCatalog *model.Catalog
}

func (s *center) Start(centerServer, endpointID, authToken string) (string, string, bool) {
	s.httpClient = &http.Client{}
	s.baseURL = fmt.Sprintf("http://%s", centerServer)

	authToken, sessionID, ok := s.verify(endpointID, authToken)
	if !ok {
		return "", "", false
	}

	log.Print("start centerAgent ok")
	return authToken, sessionID, true
}

func (s *center) Stop() {

}

func (s *center) verify(endpointID, authToken string) (string, string, bool) {
	param := &common_def.LoginEndpointParam{IdentifyID: endpointID, AuthToken: authToken}
	result := &common_def.LoginEndpointResult{}
	url := fmt.Sprintf("%s/%s", s.baseURL, "cas/endpoint/")
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("verify endpoint failed, err:%s", err.Error())
		return "", "", false
	}

	if result.ErrorCode == common_def.Success {
		return result.AuthToken, result.SessionID, true
	}

	log.Printf("verify endpoint failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return "", "", false
}

func (s *center) StrictCatalog(catalog *model.Catalog) {
	s.strictCatalog = catalog
}

func (s *center) UnstrictCatalog() {
	s.strictCatalog = nil
}
