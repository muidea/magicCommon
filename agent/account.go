package agent

import (
	"fmt"
	"log"

	common_def "muidea.com/magicCommon/def"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) LoginAccount(account, password string) (model.OnlineEntryView, string, string, bool) {
	param := &common_def.LoginAccountParam{Account: account, Password: password}
	result := &common_def.LoginAccountResult{}
	url := fmt.Sprintf("%s/%s", s.baseURL, "cas/user/")
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("login failed, err:%s", err.Error())
		return result.OnlineEntry, "", "", false
	}

	if result.ErrorCode == common_def.Success {
		return result.OnlineEntry, result.AuthToken, result.SessionID, true
	}

	return result.OnlineEntry, "", "", false
}

func (s *center) LogoutAccount(authToken, sessionID string) bool {
	if len(authToken) == 0 || len(sessionID) == 0 {
		log.Print("illegal authToken or sessionID")
		return false
	}

	result := &common_def.LogoutAccountResult{}
	url := fmt.Sprintf("%s/%s/?authToken=%s&sessionID=%s", s.baseURL, "cas/user", authToken, sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("logout failed, err:%s", err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}

func (s *center) StatusAccount(authToken, sessionID string) (model.OnlineEntryView, string, string, bool) {
	result := &common_def.StatusAccountResult{}
	if len(authToken) == 0 || len(sessionID) == 0 {
		log.Print("illegal authToken or sessionID")
		return result.OnlineEntry, "", "", false
	}

	url := fmt.Sprintf("%s/%s/?authToken=%s&sessionID=%s", s.baseURL, "cas/user", authToken, sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("get status failed, err:%s", err.Error())
		return result.OnlineEntry, "", "", false
	}

	if result.ErrorCode == common_def.Success {
		return result.OnlineEntry, result.AuthToken, result.SessionID, true
	}

	return result.OnlineEntry, "", "", false
}

func (s *center) ChangePassword(accountID int, oldPassword, newPassword, authToken, sessionID string) bool {
	param := &common_def.ChangeAccountPasswordParam{OldPassword: oldPassword, NewPassword: newPassword}
	result := &common_def.ChangeAccountPasswordResult{}
	url := fmt.Sprintf("%s/%s/%d?authToken=%s&sessionID=%s", s.baseURL, "cas/user", accountID, authToken, sessionID)
	err := net.HTTPPut(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("change account password failed, err:%s", err.Error())
		return false
	}

	if result.ErrorCode == common_def.Success {
		return true
	}

	return false
}
