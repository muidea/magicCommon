package model

import "muidea.com/magicCommon/foundation/util"

// ARTICLE 文章类型
const ARTICLE = "article"

// CATALOG 分类类型
const CATALOG = "catalog"

// LINK 链接类型
const LINK = "link"

// MEDIA 图像类型
const MEDIA = "media"

// COMMENT 注释类型
const COMMENT = "comment"

// CatalogUnit 类型单元
type CatalogUnit struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}

// IsSame 判断CatalogUnit是否相同
func (s *CatalogUnit) IsSame(r *CatalogUnit) bool {
	if r == nil {
		return false
	}

	return s.ID == r.ID && s.Type == r.Type
}

// Article 文章
type Article struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// CatalogUnit CatalogUnit
func (s *Article) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: ARTICLE}
}

// Catalog 分类
type Catalog Unit

// CatalogUnit CatalogUnit
func (s *Catalog) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: CATALOG}
}

// Link 链接
type Link Unit

// CatalogUnit CatalogUnit
func (s *Link) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: LINK}
}

// Media 文件
type Media Unit

// CatalogUnit CatalogUnit
func (s *Media) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: MEDIA}
}

// Comment 注释
type Comment struct {
	ID      int    `json:"id"`
	Subject string `json:"subject"`
}

// CatalogUnit CatalogUnit
func (s *Comment) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: COMMENT}
}

// Summary 摘要信息
type Summary struct {
	Unit
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Catalog     []CatalogUnit `json:"catalog"`
	CreateDate  string        `json:"createDate"`
	Creater     int           `json:"creater"`
}

// CatalogUnit 转换成CatalogUnit
func (s *Summary) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: s.Type}
}

// SummaryView 摘要信息显示视图
type SummaryView struct {
	Summary
	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// ArticleDetail 文章
type ArticleDetail struct {
	Article
	Catalog    []CatalogUnit `json:"catalog"`
	CreateDate string        `json:"createDate"`
	Creater    int           `json:"creater"`

	Content string `json:"content"`
}

// Summary 转换成Summary
func (s *ArticleDetail) Summary() *Summary {
	return &Summary{Unit: Unit{ID: s.ID, Name: s.Title}, Description: util.ExtractSummary(s.Content), Type: CATALOG, Catalog: s.Catalog, CreateDate: s.CreateDate, Creater: s.Creater}
}

// CatalogUnit 转换成CatalogUnit
func (s *ArticleDetail) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: ARTICLE}
}

// ArticleDetailView 文章显示信息
type ArticleDetailView struct {
	ArticleDetail
	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// CatalogDetail 分类详细信息
type CatalogDetail struct {
	Unit
	Description string        `json:"description"`
	Catalog     []CatalogUnit `json:"catalog"`
	CreateDate  string        `json:"createDate"`
	Creater     int           `json:"creater"`
}

// Summary 转换成Summary
func (s *CatalogDetail) Summary() *Summary {
	return &Summary{Unit: s.Unit, Description: s.Description, Type: CATALOG, Catalog: s.Catalog, CreateDate: s.CreateDate, Creater: s.Creater}
}

// CatalogUnit 转换成CatalogUnit
func (s *CatalogDetail) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: CATALOG}
}

// CatalogDetailView 分类详细信息显示信息
type CatalogDetailView struct {
	CatalogDetail

	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// LinkDetail 链接
type LinkDetail struct {
	Unit
	Description string        `json:"description"`
	Catalog     []CatalogUnit `json:"catalog"`
	CreateDate  string        `json:"createDate"`
	Creater     int           `json:"creater"`

	URL  string `json:"url"`
	Logo string `json:"logo"`
}

// Summary 转换成Summary
func (s *LinkDetail) Summary() *Summary {
	return &Summary{Unit: s.Unit, Description: s.Description, Type: LINK, Catalog: s.Catalog, CreateDate: s.CreateDate, Creater: s.Creater}
}

// CatalogUnit 转换成CatalogUnit
func (s *LinkDetail) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: LINK}
}

// LinkDetailView 链接显示信息
type LinkDetailView struct {
	LinkDetail

	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// MediaItem 单个文件项
type MediaItem struct {
	Name        string        `json:"name"`
	FileToken   string        `json:"fileToken"`
	Description string        `json:"description"`
	Expiration  int           `json:"expiration"`
	Catalog     []CatalogUnit `json:"catalog"`
}

// MediaDetail 文件信息
type MediaDetail struct {
	Unit
	Description string        `json:"description"`
	Catalog     []CatalogUnit `json:"catalog"`
	CreateDate  string        `json:"createDate"`
	Creater     int           `json:"creater"`

	FileToken  string `json:"fileToken"`
	Expiration int    `json:"expiration"`
}

// Summary 转换成Summary
func (s *MediaDetail) Summary() *Summary {
	return &Summary{Unit: s.Unit, Description: s.Description, Type: LINK, Catalog: s.Catalog, CreateDate: s.CreateDate, Creater: s.Creater}
}

// CatalogUnit 转换成CatalogUnit
func (s *MediaDetail) CatalogUnit() *CatalogUnit {
	return &CatalogUnit{ID: s.ID, Type: MEDIA}
}

// MediaDetailView 文件信息显示信息
type MediaDetailView struct {
	MediaDetail

	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// CommentDetail 注释
type CommentDetail struct {
	Comment
	Content    string        `json:"content"`
	Catalog    []CatalogUnit `json:"catalog"`
	CreateDate string        `json:"createDate"`
	Creater    int           `json:"creater"`
	Flag       int           `json:"flag"`
}

// Summary 转换成Summary
func (s *CommentDetail) Summary() *Summary {
	return &Summary{Unit: Unit{ID: s.ID, Name: s.Subject}, Description: util.ExtractSummary(s.Content), Type: CATALOG, Catalog: s.Catalog, CreateDate: s.CreateDate, Creater: s.Creater}
}

// CommentDetailView 注释显示信息
type CommentDetailView struct {
	CommentDetail

	Catalog []Summary `json:"catalog"`
	Creater User      `json:"creater"`
}

// ContentSummary 内容摘要信息
type ContentSummary []UnitSummary

// ContentUnit 内容项
type ContentUnit struct {
	Title      string `json:"title"`
	Type       string `json:"type"`
	CreateDate string `json:"createDate"`
}
