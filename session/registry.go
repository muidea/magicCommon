package session

import (
	"log"
	"net/http"
	"sync"
	"time"

	common_const "github.com/muidea/magicCommon/common"
	"github.com/muidea/magicCommon/foundation/util"
)

// Registry 会话仓库
type Registry interface {
	GetSession(w http.ResponseWriter, r *http.Request) Session
}

// CallBack session CallBack
type CallBack interface {
	OnTimeOut(session Session)
}

var sessionCookieID = "$session_id"

func init() {
	sessionCookieID = createUUID()
}

func createUUID() string {
	return util.RandomAlphanumeric(32)
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

// GetSession 获取Session对象
func (sm *sessionRegistryImpl) GetSession(w http.ResponseWriter, r *http.Request) Session {
	var userSession Session

	sessionID := ""
	cookie, err := r.Cookie(sessionCookieID)
	if err == nil {
		sessionID = cookie.Value
	}
	urlSession := r.URL.Query().Get(common_const.SessionID)
	if len(urlSession) > 0 {
		sessionID = urlSession
	}

	cur, found := sm.findSession(sessionID)
	if !found {
		sessionScope := r.URL.Query().Get(common_const.SessionScope)
		if sessionScope != common_const.ShareSession {
			sessionID = createUUID()
		}
		userSession = sm.createSession(sessionID)
	} else {
		userSession = cur
	}

	// 存入cookie,使用cookie存储
	sessionCookie := http.Cookie{Name: sessionCookieID, Value: userSession.ID(), Path: "/"}
	http.SetCookie(w, &sessionCookie)

	r.AddCookie(&sessionCookie)

	return userSession
}

// createSession 新建Session
func (sm *sessionRegistryImpl) createSession(sessionID string) Session {
	session := &sessionImpl{id: sessionID, context: make(map[string]interface{}), registry: sm, callBack: sm.callBack}

	session.refresh()

	sm.commandChan.insert(session)

	return session
}

func (sm *sessionRegistryImpl) findSession(sessionID string) (Session, bool) {
	session, found := sm.commandChan.find(sessionID)
	return session, found
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

func (sm *sessionRegistryImpl) insert(session *sessionImpl) {
	sm.commandChan.insert(session)
}

func (sm *sessionRegistryImpl) delete(id string) {
	sm.commandChan.remove(id)
}

func (sm *sessionRegistryImpl) find(id string) (*sessionImpl, bool) {
	return sm.commandChan.find(id)
}

func (sm *sessionRegistryImpl) count() int {
	return sm.commandChan.count()
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

type findResult struct {
	value interface{}
	found bool
}

type commandChanImpl chan commandData

func (right commandChanImpl) insert(session *sessionImpl) {
	right <- commandData{action: insert, value: session}
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

func (right commandChanImpl) find(id string) (*sessionImpl, bool) {
	reply := make(chan interface{})
	right <- commandData{action: find, value: id, result: reply}

	result := (<-reply).(*findResult)

	if result.found {
		return result.value.(*sessionImpl), result.found
	}

	return nil, false
}

func (right commandChanImpl) count() int {
	reply := make(chan interface{})
	right <- commandData{action: length, result: reply}

	result := (<-reply).(int)
	return result
}

func (right commandChanImpl) run() {
	sessionContextMap := make(map[string]*sessionImpl)
	for command := range right {
		switch command.action {
		case insert:
			session := command.value.(*sessionImpl)
			_, curOK := sessionContextMap[session.id]
			if curOK {
				log.Fatalf("duplication session id:%s", session.id)
			} else {
				sessionContextMap[session.id] = &sessionImpl{id: session.id, context: session.context, registry: session.registry, callBack: session.callBack}
			}
		case remove:
			id := command.value.(string)
			delete(sessionContextMap, id)
		case update:
			session := command.value.(*sessionImpl)
			curSesion, curOK := sessionContextMap[session.id]
			if curOK {
				curSesion.context = session.context
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
			}
			command.result <- &findResult{value: &session, found: found}
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
			command.result <- len(sessionContextMap)
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
