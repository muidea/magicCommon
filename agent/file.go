package agent

import (
	"fmt"
	"log"

	common_def "github.com/muidea/magicCommon/def"
	"github.com/muidea/magicCommon/foundation/net"
)

func (s *center) UploadFile(filePath, authToken, sessionID string) (string, bool) {
	result := &common_def.UploadFileResult{}
	fileItem := "uploadfile"
	url := fmt.Sprintf("%s/%s?key-name=%s&authToken=%s&sessionID=%s", s.baseURL, "fileregistry/file/", fileItem, authToken, sessionID)

	err := net.HTTPUpload(s.httpClient, url, fileItem, filePath, result)
	if err != nil {
		log.Printf("upload file failed, err:%s", err.Error())
		return "", false
	}

	if result.ErrorCode == common_def.Success {
		return result.FileToken, true
	}

	return "", false
}

func (s *center) DownloadFile(fileToken, filePath, authToken, sessionID string) (string, bool) {
	fileURL, ok := s.QueryFile(fileToken, authToken, sessionID)
	if !ok {
		return "", false
	}

	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, fileURL, authToken, sessionID)

	downloadFile, err := net.HTTPDownload(s.httpClient, url, filePath)
	if err != nil {
		log.Printf("download file failed, err:%s", err.Error())
		return "", false
	}

	return downloadFile, true
}

func (s *center) QueryFile(fileToken, authToken, sessionID string) (string, bool) {
	result := &common_def.DownloadFileResult{}
	url := fmt.Sprintf("%s/%s?fileToken=%s&authToken=%s&sessionID=%s", s.baseURL, "fileregistry/file/", fileToken, authToken, sessionID)

	_, err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query file failed, err:%s", err.Error())
		return result.RedirectURL, false
	}

	if result.ErrorCode == common_def.Success {
		return result.RedirectURL, true
	}

	return result.RedirectURL, false
}
