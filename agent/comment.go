package agent

import (
	"fmt"
	"log"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/model"
)

func (s *center) QueryComment(authToken, sessionID string, strictCatalog model.CatalogUnit) ([]model.CommentDetailView, bool) {
	result := &common_def.QueryCommentListResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/comments/", authToken, sessionID)

	strictStr := common_def.EncodeStrictCatalog(strictCatalog)
	if strictStr != "" {
		url = fmt.Sprintf("%s&%s", url, strictStr)
	}

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query comment failed, err:%s", err.Error())
		return result.Comment, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Comment, true
	}

	return result.Comment, false
}

func (s *center) CreateComment(subject, content string, authToken, sessionID string, strictCatalog model.CatalogUnit) (model.SummaryView, bool) {
	param := &common_def.CreateCommentParam{Subject: subject, Content: content}
	result := &common_def.CreateCommentResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "content/comment/", authToken, sessionID)

	strictStr := common_def.EncodeStrictCatalog(strictCatalog)
	if strictStr != "" {
		url = fmt.Sprintf("%s&%s", url, strictStr)
	}

	_, err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("create comment failed, err:%s", err.Error())
		return result.Comment, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Comment, true
	}

	return result.Comment, false
}

func (s *center) UpdateComment(id int, subject, content string, flag int, authToken, sessionID string) (model.SummaryView, bool) {
	param := &common_def.UpdateCommentParam{CreateCommentParam: common_def.CreateCommentParam{Subject: subject, Content: content}, Flag: flag}
	result := &common_def.UpdateCommentResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/comment", id, authToken, sessionID)

	_, err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("update comment failed, err:%s", err.Error())
		return result.Comment, false
	}

	if result.ErrorCode == common_def.Success {
		return result.Comment, true
	}

	return result.Comment, false
}

func (s *center) DeleteComment(id int, authToken, sessionID string) bool {
	result := &common_def.DestroyCommentResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "content/comment", id, authToken, sessionID)

	_, err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("delete comment failed, url:%s, err:%s", url, err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}
