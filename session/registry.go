package session

import (
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	log "github.com/cihub/seelog"
	"github.com/muidea/magicCommon/foundation/util"
)

const (
	hmacSecret = "rangh@foxmail.com"
)

func getSecret() string {
	secretVal := os.Getenv("HMAC_SECRET")
	if secretVal != "" {
		return secretVal
	}

	return hmacSecret
}

// Registry 会话仓库
type Registry interface {
	GetSession(res http.ResponseWriter, req *http.Request) Session
	CountSession(filter util.Filter) int
}

func createUUID() string {
	return util.RandomAlphanumeric(32)
}

type sessionRegistryImpl struct {
	commandChan commandChanImpl
	sessionLock sync.RWMutex
}

// CreateRegistry 创建Session仓库
func CreateRegistry() Registry {
	impl := sessionRegistryImpl{}
	impl.commandChan = make(commandChanImpl)
	go impl.commandChan.run()
	go impl.checkTimer()

	return &impl
}

// GetSession 获取Session对象
func (s *sessionRegistryImpl) GetSession(res http.ResponseWriter, req *http.Request) Session {
	sessionInfo := s.getSession(req)
	if sessionInfo != nil {
		return sessionInfo
	}

	return s.createSession(createUUID())
}

func (s *sessionRegistryImpl) CountSession(filter util.Filter) int {
	return s.count(filter)
}

func (s *sessionRegistryImpl) getSession(req *http.Request) *sessionImpl {
	authorizationValue := req.Header.Get(Authorization)
	if authorizationValue == "" {
		return nil
	}

	var sessionPtr *sessionImpl
	offset := strings.Index(authorizationValue, " ")
	if offset == -1 {
		return nil
	}

	func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorf("decode authorization failed, authorizationValue:%s, err:%v", authorizationValue, err)
			}
		}()

		s.sessionLock.Lock()
		defer s.sessionLock.Unlock()
		if authorizationValue[:offset] == jwtToken {
			sessionPtr = decodeJWT(authorizationValue[offset+1:])
		}

		if authorizationValue[:offset] == endpointToken {
			sessionPtr = decodeEndpoint(authorizationValue[offset+1:])
		}
	}()

	if sessionPtr != nil {
		curSession := s.findSession(sessionPtr.id)
		if curSession != nil {
			sessionPtr = curSession
		}

		s.sessionLock.Lock()
		defer s.sessionLock.Unlock()
		sessionPtr.registry = s
		sessionPtr.context[Authorization] = authorizationValue
	}

	return sessionPtr
}

// createSession 新建Session
func (s *sessionRegistryImpl) createSession(sessionID string) *sessionImpl {
	expiryValue := time.Now().Add(DefaultSessionTimeOutValue).UTC().Unix()
	sessionPtr := &sessionImpl{id: sessionID, context: map[string]interface{}{expiryTime: expiryValue}, observer: map[string]Observer{}, registry: s}
	sessionPtr = s.commandChan.insert(sessionPtr)

	return sessionPtr
}

func (s *sessionRegistryImpl) findSession(sessionID string) *sessionImpl {
	sessionPtr := s.commandChan.find(sessionID)
	if sessionPtr != nil {
		return sessionPtr
	}

	return nil
}

// UpdateSession 更新Session
func (s *sessionRegistryImpl) updateSession(session *sessionImpl) bool {
	return s.commandChan.update(session)
}

func (s *sessionRegistryImpl) checkTimer() {
	timeOutTimer := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-timeOutTimer.C:
			s.commandChan.checkTimeOut()
		}
	}
}

func (s *sessionRegistryImpl) insert(session *sessionImpl) *sessionImpl {
	return s.commandChan.insert(session)
}

func (s *sessionRegistryImpl) delete(id string) {
	s.commandChan.remove(id)
}

func (s *sessionRegistryImpl) find(id string) *sessionImpl {
	return s.commandChan.find(id)
}

func (s *sessionRegistryImpl) count(filter util.Filter) int {
	return s.commandChan.count(filter)
}

func (s *sessionRegistryImpl) update(session *sessionImpl) bool {
	return s.commandChan.update(session)
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
				curSession = &sessionImpl{id: session.id, context: session.context, observer: session.observer, registry: session.registry}
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
			removeList := make(map[string]*sessionImpl)
			for k, v := range sessionContextMap {
				if v.timeout() {
					removeList[k] = v
				}
			}

			for k, _ := range removeList {
				delete(sessionContextMap, k)
			}

			go func() {
				for _, v := range removeList {
					v.terminate()
				}
			}()
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

	log.Infof("session manager sessionImpl exit")
}

func (right commandChanImpl) checkTimeOut() {
	right <- commandData{action: checkTimeOut}
}
