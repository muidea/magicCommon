package def

import "muidea.com/magicCommon/model"

// QueryArticleResult 查询Article结果
type QueryArticleResult struct {
	Result
	Article model.ArticleDetailView `json:"article"`
}

// QueryArticleListResult 查询ArticleList结果
type QueryArticleListResult struct {
	Result
	Total   int                 `json:"total"`
	Article []model.SummaryView `json:"article"`
}

// CreateArticleParam 新建Article参数
type CreateArticleParam struct {
	Title   string              `json:"title"`
	Content string              `json:"content"`
	Catalog []model.CatalogUnit `json:"catalog"`
}

// CreateArticleResult 新建Article结果
type CreateArticleResult struct {
	Result
	Article model.SummaryView `json:"article"`
}

// UpdateArticleParam 更新Article参数
type UpdateArticleParam CreateArticleParam

// UpdateArticleResult 更新Article结果
type UpdateArticleResult CreateArticleResult

// DestoryArticleResult 删除Article结果
type DestoryArticleResult Result

// QueryCatalogResult 查询Catalog结果
type QueryCatalogResult struct {
	Result
	Catalog model.CatalogDetailView `json:"catalog"`
}

// QueryCatalogListResult 查询CatalogList结果
type QueryCatalogListResult struct {
	Result
	Total   int                 `json:"total"`
	Catalog []model.SummaryView `json:"catalog"`
}

// CreateCatalogParam 新建Catalog参数
type CreateCatalogParam struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Catalog     []model.CatalogUnit `json:"catalog"`
}

// CreateCatalogResult 新建Catalog结果
type CreateCatalogResult struct {
	Result
	Catalog model.SummaryView `json:"catalog"`
}

// UpdateCatalogParam 更新Catalog参数
type UpdateCatalogParam CreateCatalogParam

// UpdateCatalogResult 更新Catalog结果
type UpdateCatalogResult CreateCatalogResult

// DestroyCatalogResult 删除Catalog结果
type DestroyCatalogResult Result

// QueryLinkResult 查询Link结果
type QueryLinkResult struct {
	Result
	Link model.LinkDetailView `json:"link"`
}

// QueryLinkListResult 查询LinkList结果
type QueryLinkListResult struct {
	Result
	Total int                 `json:"total"`
	Link  []model.SummaryView `json:"link"`
}

// CreateLinkParam 新建Link参数
type CreateLinkParam struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	URL         string              `json:"url"`
	Logo        string              `json:"logo"`
	Catalog     []model.CatalogUnit `json:"catalog"`
}

// CreateLinkResult 新建Link结果
type CreateLinkResult struct {
	Result
	Link model.SummaryView `json:"link"`
}

// UpdateLinkParam 更新Link参数
type UpdateLinkParam CreateLinkParam

// UpdateLinkResult 更新Link结果
type UpdateLinkResult CreateLinkResult

// DestroyLinkResult 删除Link结果
type DestroyLinkResult Result

// QueryMediaResult 查询Media结果
type QueryMediaResult struct {
	Result
	Media model.MediaDetailView `json:"media"`
}

// QueryMediaListResult 查询MediaList结果
type QueryMediaListResult struct {
	Result
	Total int                 `json:"total"`
	Media []model.SummaryView `json:"media"`
}

// CreateMediaParam 新建Media参数
type CreateMediaParam struct {
	Name        string              `json:"name"`
	FileToken   string              `json:"fileToken"`
	Description string              `json:"description"`
	Expiration  int                 `json:"expiration"`
	Catalog     []model.CatalogUnit `json:"catalog"`
}

// CreateMediaResult 新建Media结果
type CreateMediaResult struct {
	Result
	Media model.SummaryView `json:"media"`
}

// MediaInfo MediaInfo
type MediaInfo struct {
	Name      string `json:"name"`
	FileToken string `json:"fileToken"`
}

// BatchCreateMediaParam 批量新建Media参数
type BatchCreateMediaParam struct {
	Medias      []MediaInfo         `json:"medias"`
	Description string              `json:"description"`
	Expiration  int                 `json:"expiration"`
	Catalog     []model.CatalogUnit `json:"catalog"`
}

// BatchCreateMediaResult 批量新建Media结果
type BatchCreateMediaResult struct {
	Result
	Medias []model.SummaryView `json:"media"`
}

// UpdateMediaParam 更新Media参数
type UpdateMediaParam CreateMediaParam

// UpdateMediaResult 更新Media结果
type UpdateMediaResult CreateMediaResult

// DestroyMediaResult 删除Media结果
type DestroyMediaResult Result

// CreateCommentParam 新建Comment参数
type CreateCommentParam struct {
	Subject string `json:"subject"`
	Content string `json:"description"`
}

// CreateCommentResult 新建Comment结果
type CreateCommentResult struct {
	Result
	Comment model.SummaryView `json:"comment"`
}

// UpdateCommentParam 更新Comment请求
type UpdateCommentParam struct {
	CreateCommentParam
	Flag int `json:"flag"`
}

// UpdateCommentResult 更新Comment结果
type UpdateCommentResult CreateCommentResult

// QueryCommentListResult 查询CommentList结果
type QueryCommentListResult struct {
	Result
	Total   int                       `json:"total"`
	Comment []model.CommentDetailView `json:"comment"`
}

// DestroyCommentResult 删除Comment结果
type DestroyCommentResult Result

// QuerySummaryResult 查询Summary结果
type QuerySummaryResult struct {
	Result
	Summary model.SummaryView `json:"summary"`
}

// QuerySummaryListResult 查询SummaryList结果
type QuerySummaryListResult struct {
	Result
	Total   int                 `json:"total"`
	Summary []model.SummaryView `json:"summary"`
}
