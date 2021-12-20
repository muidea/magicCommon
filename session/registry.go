package session

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/muidea/magicCommon/foundation/util"
)

// Registry 会话仓库
type Registry interface {
	GetRequestInfo(res http.ResponseWriter, req *http.Request) *SessionInfo
	GetSession(res http.ResponseWriter, req *http.Request) Session
	CountSession(filter util.Filter) int
}

// CallBack session CallBack
type CallBack interface {
	OnTimeOut(session Session)
}

var sessionCookieID = "$$session_info"

func init() {
	sessionCookieID = createUUID()
}

func createUUID() string {
	return util.RandomAlphanumeric(32)
}

func getRequestInfo(req *http.Request) *SessionInfo {
	sessionInfo := &SessionInfo{}
	if req != nil {
		sessionInfo.Decode(req)
		if sessionInfo.ID == "" || sessionInfo.Token == "" {
			cookieInfo := &SessionInfo{}
			cookie, err := req.Cookie(sessionCookieID)
			if err == nil {
				valData, valErr := base64.StdEncoding.DecodeString(cookie.Value)
				if valErr == nil {
					err = json.Unmarshal(valData, cookieInfo)
					if err == nil {
						sessionInfo = cookieInfo
					}
				}
			}
		}
	}

	return sessionInfo
}

type sessionRegistryImpl struct {
	callBack    CallBack
	commandChan commandChanImpl
	sessionLock sync.RWMutex
}

// CreateRegistry 创建Session仓库
func CreateRegistry(callback CallBack) Registry {
	impl := sessionRegistryImpl{callBack: callback}
	impl.commandChan = make(commandChanImpl)
	go impl.commandChan.run()
	go impl.checkTimer()

	return &impl
}

func (sm *sessionRegistryImpl) GetRequestInfo(res http.ResponseWriter, req *http.Request) *SessionInfo {
	return getRequestInfo(req)
}

// GetSession 获取Session对象
func (sm *sessionRegistryImpl) GetSession(res http.ResponseWriter, req *http.Request) Session {
	var userSession *sessionImpl
	sessionInfo := getRequestInfo(req)

	sessionID := sessionInfo.ID
	if sessionID == "" {
		sessionID = createUUID()
	}

	cur := sm.findSession(sessionID)
	if cur == nil {
		if sessionInfo.Scope != ShareSession {
			sessionID = createUUID()
		}
		userSession = sm.createSession(sessionID)
		sessionInfo.ID = userSession.ID()
		sessionInfo.Token = ""
	} else {
		userSession = cur
	}

	userSession.SetSessionInfo(sessionInfo)

	return userSession
}

func (sm *sessionRegistryImpl) CountSession(filter util.Filter) int {
	return sm.count(filter)
}

// createSession 新建Session
func (sm *sessionRegistryImpl) createSession(sessionID string) *sessionImpl {
	sessionPtr := &sessionImpl{id: sessionID, context: make(map[string]interface{}), registry: sm, callBack: sm.callBack}

	sessionPtr.refresh()

	sessionPtr = sm.commandChan.insert(sessionPtr)

	return sessionPtr
}

func (sm *sessionRegistryImpl) findSession(sessionID string) *sessionImpl {
	sessionPtr := sm.commandChan.find(sessionID)
	if sessionPtr != nil {
		return sessionPtr
	}

	return nil
}

// UpdateSession 更新Session
func (sm *sessionRegistryImpl) updateSession(session *sessionImpl) bool {

	return sm.commandChan.update(session)
}

func (sm *sessionRegistryImpl) checkTimer() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			sm.commandChan.checkTimeOut()
		}
	}
}

func (sm *sessionRegistryImpl) insert(session *sessionImpl) *sessionImpl {
	return sm.commandChan.insert(session)
}

func (sm *sessionRegistryImpl) delete(id string) {
	sm.commandChan.remove(id)
}

func (sm *sessionRegistryImpl) find(id string) *sessionImpl {
	return sm.commandChan.find(id)
}

func (sm *sessionRegistryImpl) count(filter util.Filter) int {
	return sm.commandChan.count(filter)
}

func (sm *sessionRegistryImpl) update(session *sessionImpl) bool {
	return sm.commandChan.update(session)
}

type commandData struct {
	action commandAction
	value  interface{}
	result chan<- interface{}
	data   chan<- map[string]interface{}
}

type commandAction int

const (
	insert commandAction = iota
	remove
	update
	find
	checkTimeOut
	length
	end
)

type commandChanImpl chan commandData

func (right commandChanImpl) insert(session *sessionImpl) *sessionImpl {
	reply := make(chan interface{})
	right <- commandData{action: insert, value: session, result: reply}

	result := (<-reply).(*sessionImpl)

	return result
}

func (right commandChanImpl) remove(id string) {
	right <- commandData{action: remove, value: id}
}

func (right commandChanImpl) update(session *sessionImpl) bool {
	reply := make(chan interface{})
	right <- commandData{action: update, value: session, result: reply}

	result := (<-reply).(bool)
	return result
}

func (right commandChanImpl) find(id string) *sessionImpl {
	reply := make(chan interface{})
	right <- commandData{action: find, value: id, result: reply}

	result := <-reply
	if result == nil {
		return nil
	}
	return result.(*sessionImpl)
}

func (right commandChanImpl) count(filter util.Filter) int {
	reply := make(chan interface{})
	right <- commandData{action: length, value: filter, result: reply}

	result := (<-reply).(int)
	return result
}

func (right commandChanImpl) run() {
	sessionContextMap := make(map[string]*sessionImpl)
	for command := range right {
		switch command.action {
		case insert:
			session := command.value.(*sessionImpl)
			curSession, curOK := sessionContextMap[session.id]
			if !curOK {
				curSession = &sessionImpl{id: session.id, context: session.context, registry: session.registry, callBack: session.callBack}
				sessionContextMap[session.id] = curSession
			}

			command.result <- curSession
		case remove:
			id := command.value.(string)
			delete(sessionContextMap, id)
		case update:
			session := command.value.(*sessionImpl)
			curSession, curOK := sessionContextMap[session.id]
			if curOK {
				curSession.context = session.context
			} else {
				log.Printf("illegal session id:%s", session.id)
			}

			command.result <- curOK
		case find:
			id := command.value.(string)
			var session sessionImpl
			cur, found := sessionContextMap[id]
			if found {
				cur.refresh()
				session = *cur
				command.result <- &session
			} else {
				command.result <- nil
			}
		case checkTimeOut:
			removeList := []string{}
			for k, v := range sessionContextMap {
				if v.timeOut() {
					if v.callBack != nil {
						go v.callBack.OnTimeOut(v)
					}

					removeList = append(removeList, k)
				}
			}

			for key := range removeList {
				delete(sessionContextMap, removeList[key])
			}
		case length:
			filter := command.value.(util.Filter)
			if filter == nil {
				command.result <- len(sessionContextMap)
				return
			}
			count := 0
			for _, val := range sessionContextMap {
				if filter.Filter(val) {
					count++
				}
			}
			command.result <- count
		case end:
			close(right)
		}
	}

	log.Print("session manager sessionImpl exit")
}

/*
func (right commandChanImpl) close() map[string]interface{} {
	reply := make(chan map[string]interface{})
	right <- commandData{action: end, data: reply}
	return <-reply
}
*/

func (right commandChanImpl) checkTimeOut() {
	right <- commandData{action: checkTimeOut}
}
