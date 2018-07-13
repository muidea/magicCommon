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
	Article []model.SummaryView `json:"article"`
}

// CreateArticleParam 新建Article参数
type CreateArticleParam struct {
	Name    string          `json:"name"`
	Content string          `json:"content"`
	Catalog []model.Catalog `json:"catalog"`
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
	Catalog []model.SummaryView `json:"catalog"`
}

// CreateCatalogParam 新建Catalog参数
type CreateCatalogParam struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Catalog     []model.Catalog `json:"catalog"`
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
	Link []model.SummaryView `json:"link"`
}

// CreateLinkParam 新建Link参数
type CreateLinkParam struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	URL         string          `json:"url"`
	Logo        string          `json:"logo"`
	Catalog     []model.Catalog `json:"catalog"`
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
	Media []model.SummaryView `json:"media"`
}

// CreateMediaParam 新建Media参数
type CreateMediaParam struct {
	Name        string          `json:"name"`
	FileToken   string          `json:"fileToken"`
	Description string          `json:"description"`
	Expiration  int             `json:"expiration"`
	Catalog     []model.Catalog `json:"catalog"`
}

// CreateMediaResult 新建Media结果
type CreateMediaResult struct {
	Result
	Media model.SummaryView `json:"media"`
}

// MediaItem mediaItem
type MediaItem struct {
	Name      string `json:"name"`
	FileToken string `json:"fileToken"`
}

// BatchCreateMediaParam 批量新建Media参数
type BatchCreateMediaParam struct {
	Medias      []MediaItem     `json:"medias"`
	Description string          `json:"description"`
	Expiration  int             `json:"expiration"`
	Catalog     []model.Catalog `json:"catalog"`
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

// QuerySummaryResult 查询Summary结果
type QuerySummaryResult struct {
	Result
	Summary model.SummaryView `json:"summary"`
}

// QuerySummaryListResult 查询SummaryList结果
type QuerySummaryListResult struct {
	Result
	Summary []model.SummaryView `json:"summary"`
}