package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) QuerySyslog(source string, filter *common_def.PageFilter, sessionToken, sessionID string) ([]model.Syslog, int) {
	result := &common_def.QuerySyslogResult{}

	url := fmt.Sprintf("%s/%s/?authToken=%s&sessionID=%s&source=%s", s.baseURL, "system/syslog", sessionToken, sessionID, source)
	if filter != nil {
		filterVal := filter.Encode()
		if filterVal != "" {
			url = fmt.Sprintf("%s&%s", url, filterVal)
		}
	}

	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("query syslog failed, err:%s", err.Error())
		return result.Syslog, result.Total
	}

	if result.ErrorCode == common_def.Success {
		return result.Syslog, result.Total
	}

	return result.Syslog, result.Total
}

func (s *center) InsertSyslog(user, operation, datetime, source, sessionToken, sessionID string) bool {
	param := &common_def.InsertSyslogParam{User: user, Operation: operation, DateTime: datetime, Source: source}
	result := &common_def.InsertSyslogResult{}
	url := fmt.Sprintf("%s/%s?authToken=%s&sessionID=%s", s.baseURL, "system/syslog/", sessionToken, sessionID)

	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("insert syslog failed, err:%s", err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}
