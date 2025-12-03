package session

import (
	"context"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/muidea/magicCommon/foundation/log"
	fn "github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/foundation/util"
)

const (
	hmacSecretDefault = "rangh@foxmail.com"
	HMAC_SECRET_KEY   = "HMAC_SECRET"
)

func getSecret() string {
	secretVal := os.Getenv("HMAC_SECRET")
	if secretVal != "" {
		return secretVal
	}

	return hmacSecretDefault
}

// Registry 会话仓库
type Registry interface {
	GetSession(res http.ResponseWriter, req *http.Request) Session
	CountSession(filter util.Filter) int
	Release()
}

func createUUID() string {
	return util.RandomAlphanumeric(32)
}

type sessionRegistryImpl struct {
	commandChan        commandChanImpl
	sessionLock        sync.RWMutex
	registryCancelFunc context.CancelFunc
}

// CreateRegistry 创建Session仓库
func CreateRegistry() Registry {
	registryCtx, registryCancel := context.WithCancel(context.Background())
	impl := sessionRegistryImpl{}
	impl.registryCancelFunc = registryCancel
	impl.commandChan = make(commandChanImpl)
	go impl.commandChan.run()
	go impl.checkTimer(registryCtx)

	return &impl
}

// GetSession 获取Session对象
func (s *sessionRegistryImpl) GetSession(res http.ResponseWriter, req *http.Request) Session {
	sessionInfo := s.getSession(req)
	if sessionInfo != nil {
		return sessionInfo
	}

	return s.createSession(req, createUUID())
}

func (s *sessionRegistryImpl) Release() {
	s.registryCancelFunc()
	s.commandChan.end()
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

		nowTime := time.Now().UTC().UnixMilli()
		s.sessionLock.Lock()
		defer s.sessionLock.Unlock()
		if authorizationValue[:offset] == jwtToken {
			sessionPtr = decodeJWT(authorizationValue[offset+1:])
		}

		if authorizationValue[:offset] == sigToken {
			sessionPtr = decodeEndpoint(authorizationValue[offset+1:])
		}
		if sessionPtr != nil {
			// 到这里说明session已经失效了，
			expireTime := sessionPtr.getExpireTime()
			if expireTime < nowTime {
				sessionPtr = nil
			}
		}
	}()

	if sessionPtr != nil {
		sessionPtr.context[InnerRemoteAccessAddr] = fn.GetHTTPRemoteAddress(req)
		sessionPtr.context[InnerUseAgent] = req.UserAgent()

		curSession := s.findSession(sessionPtr.id)
		if curSession != nil {
			// 到这里说明当前会话失效了。需要重新认证
			if curSession.isFinal() {
				return nil
			}
			sessionPtr = curSession
		} else {
			sessionPtr = s.insertSession(sessionPtr)
		}

		s.sessionLock.Lock()
		defer s.sessionLock.Unlock()
		sessionPtr.context[Authorization] = authorizationValue
	}

	return sessionPtr
}

// createSession 新建Session
func (s *sessionRegistryImpl) createSession(req *http.Request, sessionID string) *sessionImpl {
	expireValue := time.Now().Add(DefaultSessionTimeOutValue).UTC().UnixMilli()
	sessionPtr := &sessionImpl{id: sessionID, context: map[string]any{innerExpireTime: expireValue}, observer: map[string]Observer{}, registry: s}
	sessionPtr.context[InnerRemoteAccessAddr] = fn.GetHTTPRemoteAddress(req)
	sessionPtr.context[InnerUseAgent] = req.UserAgent()
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

func (s *sessionRegistryImpl) insertSession(sessionPtr *sessionImpl) *sessionImpl {
	sessionPtr.registry = s
	return s.commandChan.insert(sessionPtr)
}

// UpdateSession 更新Session
func (s *sessionRegistryImpl) updateSession(sessionPtr *sessionImpl) bool {
	return s.commandChan.update(sessionPtr)
}

func (s *sessionRegistryImpl) checkTimer(ctx context.Context) {
	timeOutTimer := time.NewTicker(5 * time.Second)
	defer timeOutTimer.Stop() // 确保在函数退出时停止定时器

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeOutTimer.C:
			s.commandChan.checkTimeOut()
		}
	}
}

func (s *sessionRegistryImpl) count(filter util.Filter) int {
	return s.commandChan.count(filter)
}

type commandData struct {
	action commandAction
	value  interface{}
	result chan<- interface{}
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
	defer close(reply)
	right <- commandData{action: insert, value: session, result: reply}

	result := (<-reply).(*sessionImpl)

	return result
}

func (right commandChanImpl) update(session *sessionImpl) bool {
	reply := make(chan interface{})
	defer close(reply)
	right <- commandData{action: update, value: session, result: reply}

	result := (<-reply).(bool)
	return result
}

func (right commandChanImpl) find(id string) *sessionImpl {
	reply := make(chan interface{})
	defer close(reply)

	right <- commandData{action: find, value: id, result: reply}

	result := <-reply
	if result == nil {
		return nil
	}
	return result.(*sessionImpl)
}

func (right commandChanImpl) count(filter util.Filter) int {
	reply := make(chan interface{})
	defer close(reply)
	right <- commandData{action: length, value: filter, result: reply}

	result := (<-reply).(int)
	return result
}

func (right commandChanImpl) end() {
	result := make(chan interface{})
	defer close(result)
	right <- commandData{action: end, result: result}
	<-result
}

func (right commandChanImpl) run() {
	sessionContextMap := make(map[string]*sessionImpl)
	for command := range right {
		switch command.action {
		case insert:
			session := command.value.(*sessionImpl)
			curSession, curOK := sessionContextMap[session.id]
			if !curOK {
				curSession = &sessionImpl{
					id:      session.id,
					context: session.context, observer: session.observer, registry: session.registry}
				curSession.context[InnerStartTime] = time.Now().UTC().UnixMilli()
				sessionContextMap[session.id] = curSession
			}
			curSession.refresh()
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
			var session *sessionImpl
			cur, found := sessionContextMap[id]
			if found {
				if !cur.timeout() {
					cur.refresh()
				}
				session = cur
				command.result <- session
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

			for k := range removeList {
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
			command.result <- true
			close(right)
		}
	}

	log.Infof("session manager sessionImpl exit")
}

func (right commandChanImpl) checkTimeOut() {
	right <- commandData{action: checkTimeOut}
}
