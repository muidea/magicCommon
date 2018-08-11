package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
)

func (s *center) UploadFile(filePath, authToken, sessionID string) (string, bool) {
	return "", true
}

func (s *center) DownloadFile(fileToken, authToken, sessionID string) (string, bool) {
	return "", true
}

func (s *center) QueryFile(fileToken, authToken, sessionID string) (string, bool) {
	result := &common_def.DownloadFileResult{}
	url := fmt.Sprintf("%s/%s?fileToken=%s&authToken=%s&sessionID=%s", s.baseURL, "fileregistry/file/", fileToken, authToken, sessionID)

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query file failed, err:%s", err.Error())
		return result.RedirectURL, false
	}

	if result.ErrorCode == common_def.Success {
		return result.RedirectURL, true
	}

	log.Printf("query file failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.RedirectURL, false
}
