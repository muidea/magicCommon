package agent

import (
	"fmt"
	"log"
	"net/http"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/foundation/util"
	"github.com/muidea/magicCommon/model"
)

// Agent Center访问代理
type Agent interface {
	// @return
	// agent sessionToken
	// agent sessionID
	// result(true or false)
	Start(centerServer, endpointID, endpointToken string) (string, string, bool)
	Stop()

	// @return
	// account sessionToken
	// account sessionID
	// result(true or false)
	LoginAccount(account, password string) (model.OnlineEntryView, string, string, bool)
	LogoutAccount(sessionToken, sessionID string) bool
	// @return
	// account entryView
	// account sessionToken
	// account sessionID
	// result(true or false)
	StatusAccount(sessionToken, sessionID string) (model.OnlineEntryView, string, string, bool)
	ChangePassword(accountID int, oldPassword, newPassword, sessionToken, sessionID string) bool

	FetchSummary(name, summaryType, sessionToken, sessionID string, strictCatalog *model.CatalogUnit) (model.SummaryView, bool)
	QuerySummaryContent(summary model.CatalogUnit, filter *common_def.Filter, sessionToken, sessionID string) ([]model.SummaryView, int)
	QuerySummaryContentWithSpecialType(summary model.CatalogUnit, specialType string, filter *common_def.Filter, sessionToken, sessionID string) ([]model.SummaryView, int)
	QuerySummaryContentByUser(user int, filter *common_def.Filter, sessionToken, sessionID string, strictCatalog *model.CatalogUnit) ([]model.SummaryView, int)

	QueryCatalog(id int, sessionToken, sessionID string) (model.CatalogDetailView, bool)
	CreateCatalog(name, description string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	UpdateCatalog(id int, name, description string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	DeleteCatalog(id int, sessionToken, sessionID string) bool

	QueryArticle(id int, sessionToken, sessionID string) (model.ArticleDetailView, bool)
	CreateArticle(title, content string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	UpdateArticle(id int, title, content string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	DeleteArticle(id int, sessionToken, sessionID string) bool

	QueryLink(id int, sessionToken, sessionID string) (model.LinkDetailView, bool)
	CreateLink(name, description, url, logo string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	UpdateLink(id int, name, description, url, logo string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	DeleteLink(id int, sessionToken, sessionID string) bool

	QueryMedia(id int, sessionToken, sessionID string) (model.MediaDetailView, bool)
	CreateMedia(name, description, fileToken string, expiration int, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	BatchCreateMedia(media []common_def.MediaInfo, description string, catalog []model.CatalogUnit, expiration int, sessionToken, sessionID string) ([]model.SummaryView, bool)
	UpdateMedia(id int, name, description, fileToken string, expiration int, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool)
	DeleteMedia(id int, sessionToken, sessionID string) bool

	QueryComment(sessionToken, sessionID string, strictCatalog model.CatalogUnit) ([]model.CommentDetailView, bool)
	CreateComment(subject, content string, sessionToken, sessionID string, strictCatalog model.CatalogUnit) (model.SummaryView, bool)
	UpdateComment(id int, subject, content string, flag int, sessionToken, sessionID string) (model.SummaryView, bool)
	DeleteComment(id int, sessionToken, sessionID string) bool

	UploadFile(filePath, sessionToken, sessionID string) (string, bool)
	DownloadFile(fileToken, filePath, sessionToken, sessionID string) (string, bool)
	QueryFile(fileToken, sessionToken, sessionID string) (string, bool)

	QuerySyslog(source string, filter *util.PageFilter, sessionToken, sessionID string) ([]model.Syslog, int)
	InsertSyslog(user, operation, datetime, source, sessionToken, sessionID string) bool
}

// New 新建Agent
func New() Agent {
	return &center{}
}

type center struct {
	httpClient *http.Client
	baseURL    string
}

func (s *center) Start(centerServer, endpointID, endpointToken string) (string, string, bool) {
	s.httpClient = &http.Client{}
	s.baseURL = fmt.Sprintf("http://%s", centerServer)

	sessionToken, sessionID, ok := s.verify(endpointID, endpointToken)
	if !ok {
		return "", "", false
	}

	return sessionToken, sessionID, true
}

func (s *center) Stop() {

}

func (s *center) verify(endpointID, endpointToken string) (string, string, bool) {
	param := &common_def.LoginEndpointParam{IdentifyID: endpointID, AuthToken: endpointToken}
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

	return "", "", false
}
