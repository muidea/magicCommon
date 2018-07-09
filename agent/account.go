package agent

import (
	"fmt"
	"log"

	common_result "muidea.com/magicCommon/common"
	"muidea.com/magicCommon/foundation/net"
	"muidea.com/magicCommon/model"
)

func (s *center) LoginAccount(account, password string) (model.AccountOnlineView, string, string, bool) {
	type loginParam struct {
		Account  string `json:"account"`
		Password string `json:"password"`
	}

	type loginResult struct {
		common_result.Result
		OnlineUser model.AccountOnlineView `json:"onlineUser"`
		SessionID  string                  `json:"sessionID"`
	}

	param := &loginParam{Account: account, Password: password}
	result := &loginResult{}
	url := fmt.Sprintf("%s/%s", s.baseURL, "cas/user/")
	err := net.HTTPPost(s.httpClient, url, param, result)
	if err != nil {
		log.Printf("login failed, err:%s", err.Error())
		return result.OnlineUser, "", "", false
	}

	if result.ErrorCode == common_result.Success {
		return result.OnlineUser, result.OnlineUser.AuthToken, result.SessionID, true
	}

	log.Printf("login failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.OnlineUser, "", "", false
}

func (s *center) LogoutAccount(authToken, sessionID string) bool {
	type logoutResult struct {
		common_result.Result
	}

	if len(authToken) == 0 || len(sessionID) == 0 {
		log.Print("illegal authToken or sessionID")
		return false
	}

	result := &logoutResult{}
	url := fmt.Sprintf("%s/%s/?authToken=%s&sessionID=%s", s.baseURL, "cas/user", authToken, sessionID)
	err := net.HTTPDelete(s.httpClient, url, result)
	if err != nil {
		log.Printf("logout failed, err:%s", err.Error())
		return false
	}

	if result.ErrorCode == common_result.Success {
		return true
	}

	log.Printf("logout failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return false
}

func (s *center) StatusAccount(authToken, sessionID string) (model.AccountOnlineView, bool) {
	type statusResult struct {
		common_result.Result
		OnlineUser model.AccountOnlineView `json:"onlineUser"`
		SessionID  string                  `json:"sessionID"`
	}

	result := &statusResult{}
	if len(authToken) == 0 || len(sessionID) == 0 {
		log.Print("illegal authToken or sessionID")
		return result.OnlineUser, false
	}

	url := fmt.Sprintf("%s/%s/?authToken=%s&sessionID=%s", s.baseURL, "cas/user", authToken, sessionID)
	err := net.HTTPGet(s.httpClient, url, result)
	if err != nil {
		log.Printf("get status failed, err:%s", err.Error())
		return result.OnlineUser, false
	}

	if result.ErrorCode == common_result.Success {
		return result.OnlineUser, true
	}

	log.Printf("status account failed, errorCode:%d, reason:%s", result.ErrorCode, result.Reason)
	return result.OnlineUser, false
}
