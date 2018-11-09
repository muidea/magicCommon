package agent

import (
	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/model"
)

// Wrapper agent wrapper
type Wrapper struct {
	agent        Agent
	sessionToken string
	sessionID    string
}

//NewWrapper new Wrapper
func NewWrapper() *Wrapper {
	return &Wrapper{agent: New()}
}

// Start start agent
func (s *Wrapper) Start(centerSvr, endpointID, endpointToken string) bool {
	sessionToken, sessionID, ok := s.agent.Start(centerSvr, endpointID, endpointToken)
	if ok {
		s.sessionToken = sessionToken
		s.sessionID = sessionID
		return ok
	}
	return false
}

// Stop stop agent
func (s *Wrapper) Stop() {
	s.agent.Stop()
	s.sessionToken = ""
	s.sessionID = ""
}

// SessionToken 获取sessionToken
func (s *Wrapper) SessionToken() string {
	return s.sessionToken
}

// SessionID 获取sessionID
func (s *Wrapper) SessionID() string {
	return s.sessionID
}

// LoginAccount login account
func (s *Wrapper) LoginAccount(account, password string) (model.OnlineEntryView, string, string, bool) {
	return s.agent.LoginAccount(account, password)
}

// LogoutAccount logout account
func (s *Wrapper) LogoutAccount(sessionToken, sessionID string) bool {
	return s.agent.LogoutAccount(sessionToken, sessionID)
}

// StatusAccount status account
func (s *Wrapper) StatusAccount(sessionToken, sessionID string) (model.OnlineEntryView, string, string, bool) {
	return s.agent.StatusAccount(sessionToken, sessionID)
}

// ChangePassword change password
func (s *Wrapper) ChangePassword(accountID int, oldPassword, newPassword, sessionToken, sessionID string) bool {
	return s.agent.ChangePassword(accountID, oldPassword, newPassword, sessionToken, sessionID)
}

// FetchSummary fetech summary
func (s *Wrapper) FetchSummary(name, summaryType, sessionToken, sessionID string, strictCatalog *model.CatalogUnit) (model.SummaryView, bool) {
	return s.agent.FetchSummary(name, summaryType, sessionToken, sessionID, strictCatalog)
}

// QuerySummaryContent query summaryContent
func (s *Wrapper) QuerySummaryContent(summary model.CatalogUnit, filter *common_def.Filter, sessionToken, sessionID string) ([]model.SummaryView, int) {
	return s.agent.QuerySummaryContent(summary, filter, sessionToken, sessionID)
}

// QuerySummaryContentWithSpecialType query SummaryContent with specialType
func (s *Wrapper) QuerySummaryContentWithSpecialType(summary model.CatalogUnit, specialType string, filter *common_def.Filter, sessionToken, sessionID string) ([]model.SummaryView, int) {
	return s.agent.QuerySummaryContentWithSpecialType(summary, specialType, filter, sessionToken, sessionID)
}

// QuerySummaryContentByUser query SummaryContent by User
func (s *Wrapper) QuerySummaryContentByUser(user int, filter *common_def.Filter, sessionToken, sessionID string, strictCatalog *model.CatalogUnit) ([]model.SummaryView, int) {
	return s.agent.QuerySummaryContentByUser(user, filter, sessionToken, sessionID, strictCatalog)
}

// QueryCatalog query catalog
func (s *Wrapper) QueryCatalog(id int, sessionToken, sessionID string) (model.CatalogDetailView, bool) {
	return s.agent.QueryCatalog(id, sessionToken, sessionID)
}

// CreateCatalog create catalog
func (s *Wrapper) CreateCatalog(name, description string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.CreateCatalog(name, description, catalog, sessionToken, sessionID)
}

// UpdateCatalog update catalog
func (s *Wrapper) UpdateCatalog(id int, name, description string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.UpdateCatalog(id, name, description, catalog, sessionToken, sessionID)
}

// DeleteCatalog delete catalog
func (s *Wrapper) DeleteCatalog(id int, sessionToken, sessionID string) bool {
	return s.agent.DeleteCatalog(id, sessionToken, sessionID)
}

// QueryArticle query article
func (s *Wrapper) QueryArticle(id int, sessionToken, sessionID string) (model.ArticleDetailView, bool) {
	return s.agent.QueryArticle(id, sessionToken, sessionID)
}

// CreateArticle create article
func (s *Wrapper) CreateArticle(title, content string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.CreateArticle(title, content, catalog, sessionToken, sessionID)
}

// UpdateArticle update article
func (s *Wrapper) UpdateArticle(id int, title, content string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.UpdateArticle(id, title, content, catalog, sessionToken, sessionID)
}

// DeleteArticle delete article
func (s *Wrapper) DeleteArticle(id int, sessionToken, sessionID string) bool {
	return s.agent.DeleteArticle(id, sessionToken, sessionID)
}

// QueryLink query link
func (s *Wrapper) QueryLink(id int, sessionToken, sessionID string) (model.LinkDetailView, bool) {
	return s.agent.QueryLink(id, sessionToken, sessionID)
}

// CreateLink create link
func (s *Wrapper) CreateLink(name, description, url, logo string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.CreateLink(name, description, url, logo, catalog, sessionToken, sessionID)
}

// UpdateLink update link
func (s *Wrapper) UpdateLink(id int, name, description, url, logo string, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.UpdateLink(id, name, description, url, logo, catalog, sessionToken, sessionID)
}

// DeleteLink delete link
func (s *Wrapper) DeleteLink(id int, sessionToken, sessionID string) bool {
	return s.agent.DeleteLink(id, sessionToken, sessionID)
}

// QueryMedia query media
func (s *Wrapper) QueryMedia(id int, sessionToken, sessionID string) (model.MediaDetailView, bool) {
	return s.agent.QueryMedia(id, sessionToken, sessionID)
}

// CreateMedia create media
func (s *Wrapper) CreateMedia(name, description, fileToken string, expiration int, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.CreateMedia(name, description, fileToken, expiration, catalog, sessionToken, sessionID)
}

// BatchCreateMedia batch create meia
func (s *Wrapper) BatchCreateMedia(media []common_def.MediaInfo, description string, catalog []model.CatalogUnit, expiration int, sessionToken, sessionID string) ([]model.SummaryView, bool) {
	return s.agent.BatchCreateMedia(media, description, catalog, expiration, sessionToken, sessionID)
}

// UpdateMedia update media
func (s *Wrapper) UpdateMedia(id int, name, description, fileToken string, expiration int, catalog []model.CatalogUnit, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.UpdateMedia(id, name, description, fileToken, expiration, catalog, sessionToken, sessionID)
}

// DeleteMedia delete media
func (s *Wrapper) DeleteMedia(id int, sessionToken, sessionID string) bool {
	return s.agent.DeleteMedia(id, sessionToken, sessionID)
}

// QueryComment query comment
func (s *Wrapper) QueryComment(sessionToken, sessionID string, strictCatalog model.CatalogUnit) ([]model.CommentDetailView, bool) {
	return s.agent.QueryComment(sessionToken, sessionID, strictCatalog)
}

// CreateComment create comment
func (s *Wrapper) CreateComment(subject, content string, sessionToken, sessionID string, strictCatalog model.CatalogUnit) (model.SummaryView, bool) {
	return s.agent.CreateComment(subject, content, sessionToken, sessionID, strictCatalog)
}

// UpdateComment update comment
func (s *Wrapper) UpdateComment(id int, subject, content string, flag int, sessionToken, sessionID string) (model.SummaryView, bool) {
	return s.agent.UpdateComment(id, subject, content, flag, sessionToken, sessionID)
}

// DeleteComment delete comment
func (s *Wrapper) DeleteComment(id int, sessionToken, sessionID string) bool {
	return s.agent.DeleteComment(id, sessionToken, sessionID)
}

// UploadFile upload file
func (s *Wrapper) UploadFile(filePath, sessionToken, sessionID string) (string, bool) {
	return s.agent.UploadFile(filePath, sessionToken, sessionID)
}

// DownloadFile download file
func (s *Wrapper) DownloadFile(fileToken, filePath, sessionToken, sessionID string) (string, bool) {
	return s.agent.DownloadFile(fileToken, filePath, sessionToken, sessionID)
}

// QueryFile query file
func (s *Wrapper) QueryFile(fileToken, sessionToken, sessionID string) (string, bool) {
	return s.agent.QueryFile(fileToken, sessionToken, sessionID)
}

// QuerySyslog query syslog
func (s *Wrapper) QuerySyslog(source string, filter *common_def.PageFilter, sessionToken, sessionID string) ([]model.Syslog, int) {
	return s.agent.QuerySyslog(source, filter, sessionToken, sessionID)
}

// InsertSyslog insert syslog
func (s *Wrapper) InsertSyslog(user, operation, datetime, source, sessionToken, sessionID string) bool {
	return s.agent.InsertSyslog(user, operation, datetime, source, sessionToken, sessionID)
}
