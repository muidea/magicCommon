package session

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	fn "github.com/muidea/magicCommon/foundation/net"
	"github.com/muidea/magicCommon/foundation/util"
	"log/slog"
)

const (
	SESSION_TIMEOUT_VALUE_KEY = "SESSION_TIMEOUT_VALUE"
)

func GetSessionTimeOutValue() time.Duration {
	sessionTimeoutVal := os.Getenv(SESSION_TIMEOUT_VALUE_KEY)
	if sessionTimeoutVal != "" {
		iVal, iErr := strconv.Atoi(sessionTimeoutVal)
		if iErr != nil || iVal <= 0 {
			return DefaultSessionTimeOutValue
		}

		return time.Duration(iVal) * time.Minute
	}

	return DefaultSessionTimeOutValue
}

// Registry 会话仓库
type Registry interface {
	GetSession(res http.ResponseWriter, req *http.Request) Session
	CountSession(filter util.Filter) int
	Release()
}

func LookupSession(reg Registry, req *http.Request) Session {
	if reg == nil || req == nil {
		return nil
	}

	impl, ok := reg.(*sessionRegistryImpl)
	if !ok {
		return nil
	}

	curSession := impl.getSession(req)
	if curSession == nil {
		return nil
	}
	return curSession
}

func ResolveSession(reg Registry, req *http.Request) Session {
	if req == nil {
		return nil
	}
	if curSession := LookupSession(reg, req); curSession != nil {
		return curSession
	}
	return NewAnonymousSession(fn.GetHTTPRemoteAddress(req), req.UserAgent())
}

func createUUID() string {
	return util.RandomAlphanumeric(32)
}

type sessionRegistryImpl struct {
	registryCancelFunc context.CancelFunc
	registryLock       sync.RWMutex
	sessionMap         map[string]*sessionImpl
	sessionObserver    Observer
	releaseOnce        sync.Once
}

// DefaultRegistry 创建Session仓库
func DefaultRegistry() Registry {
	return NewRegistry(nil)
}

func NewRegistry(obSvr Observer) Registry {
	registryCtx, registryCancel := context.WithCancel(context.Background())
	impl := sessionRegistryImpl{
		sessionObserver: obSvr,
		sessionMap:      map[string]*sessionImpl{},
	}
	impl.registryCancelFunc = registryCancel
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
	s.releaseOnce.Do(func() {
		s.registryCancelFunc()
	})
}

func (s *sessionRegistryImpl) CountSession(filter util.Filter) int {
	return s.count(filter)
}

func (s *sessionRegistryImpl) getSession(req *http.Request) *sessionImpl {
	// Phase-1 auth authority is CAS OAuth/OIDC (JWKS-verified access tokens).
	// Registry is only an in-process opaque session container:
	// - Cookie value is an opaque session id (never JWT claims material).
	// - Authorization: Sig is rejected (Endpoint AES tokens are retired).
	// - Authorization: Bearer is NOT decoded here; resource services verify JWT via JWKS.
	var sessionPtr *sessionImpl
	func() {
		defer func() {
			if err := recover(); err != nil {
				sessionPtr = nil
				stackInfo := util.GetStack(3)
				slog.Error("get session failed, err:err, stack:\nstackInfo", "field", err, "error", stackInfo)
			}
		}()

		// Explicitly refuse the retired Endpoint Sig scheme. This is a
		// fail-closed parser branch, not a public authentication API.
		scheme, _ := ParseAuthorizationScheme(req.Header.Get(Authorization))
		if strings.EqualFold(scheme, "Sig") {
			return
		}

		sessionID := strings.TrimSpace(ReadSessionTokenFromCookie(req))
		if sessionID == "" {
			return
		}
		// Opaque id only: never decode legacy JWT or Endpoint tokens from cookie.
		// Reject values that look like JWTs (three base64url segments).
		if looksLikeJWT(sessionID) {
			return
		}

		curSession := s.findSession(sessionID)
		if curSession == nil || curSession.isFinal() {
			if curSession != nil && curSession.isFinal() {
				s.removeSession(curSession.id)
			}
			return
		}
		if curSession.timeout() {
			s.removeSession(curSession.id)
			return
		}
		curSession.refresh()
		curSession.mu.Lock()
		curSession.context[InnerRemoteAccessAddr] = fn.GetHTTPRemoteAddress(req)
		curSession.context[InnerUseAgent] = req.UserAgent()
		curSession.mu.Unlock()
		sessionPtr = curSession
	}()

	return sessionPtr
}

// looksLikeJWT reports compact JWS shape (header.payload.signature).
func looksLikeJWT(val string) bool {
	if val == "" {
		return false
	}
	parts := strings.Split(val, ".")
	return len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != ""
}

// createSession 新建Session
func (s *sessionRegistryImpl) createSession(req *http.Request, sessionID string) *sessionImpl {
	expireValue := time.Now().Add(GetSessionTimeOutValue()).UTC().UnixMilli()
	sessionPtr := &sessionImpl{id: sessionID, context: map[string]any{innerExpireTime: expireValue}, observer: map[string]Observer{}, registry: s}
	sessionPtr.context[InnerRemoteAccessAddr] = fn.GetHTTPRemoteAddress(req)
	sessionPtr.context[InnerUseAgent] = req.UserAgent()
	sessionPtr = s.insertSession(sessionPtr)
	if s.sessionObserver != nil {
		sessionPtr.mu.Lock()
		sessionPtr.observer[s.sessionObserver.ID()] = s.sessionObserver
		sessionPtr.mu.Unlock()
	}

	return sessionPtr
}

func (s *sessionRegistryImpl) findSession(sessionID string) *sessionImpl {
	s.registryLock.RLock()
	sessionPtr := s.sessionMap[sessionID]
	s.registryLock.RUnlock()
	return sessionPtr
}

func (s *sessionRegistryImpl) insertSession(sessionPtr *sessionImpl) *sessionImpl {
	sessionPtr.registry = s
	s.registryLock.Lock()
	defer s.registryLock.Unlock()

	curSession, curOK := s.sessionMap[sessionPtr.id]
	if !curOK {
		curSession = &sessionImpl{
			id:       sessionPtr.id,
			context:  sessionPtr.context,
			observer: sessionPtr.observer,
			registry: sessionPtr.registry,
			status:   sessionPtr.status,
		}
		curSession.context[InnerStartTime] = time.Now().UTC().UnixMilli()
		s.sessionMap[sessionPtr.id] = curSession
	}
	curSession.refresh()
	return curSession
}

// UpdateSession 更新Session
func (s *sessionRegistryImpl) updateSession(sessionPtr *sessionImpl) bool {
	if s == nil || sessionPtr == nil {
		return false
	}

	s.registryLock.RLock()
	curSession, curOK := s.sessionMap[sessionPtr.id]
	s.registryLock.RUnlock()
	if !curOK {
		return false
	}

	if curSession == sessionPtr {
		return true
	}

	sessionPtr.mu.RLock()
	contextCopy := make(map[string]any, len(sessionPtr.context))
	for k, v := range sessionPtr.context {
		contextCopy[k] = v
	}
	observerCopy := make(map[string]Observer, len(sessionPtr.observer))
	for k, v := range sessionPtr.observer {
		observerCopy[k] = v
	}
	statusVal := sessionPtr.status
	sessionPtr.mu.RUnlock()

	curSession.mu.Lock()
	curSession.context = contextCopy
	curSession.observer = observerCopy
	curSession.status = statusVal
	curSession.mu.Unlock()
	return true
}

func (s *sessionRegistryImpl) checkTimer(ctx context.Context) {
	timeOutTimer := time.NewTicker(5 * time.Second)
	defer timeOutTimer.Stop() // 确保在函数退出时停止定时器

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeOutTimer.C:
			var removeList []*sessionImpl

			s.registryLock.Lock()
			for id, val := range s.sessionMap {
				if val.timeout() {
					removeList = append(removeList, val)
					delete(s.sessionMap, id)
				}
			}
			s.registryLock.Unlock()

			for _, val := range removeList {
				go val.terminate()
			}
		}
	}
}

func (s *sessionRegistryImpl) count(filter util.Filter) int {
	s.registryLock.RLock()
	defer s.registryLock.RUnlock()

	if filter == nil {
		return len(s.sessionMap)
	}

	count := 0
	for _, val := range s.sessionMap {
		if filter.Filter(val) {
			count++
		}
	}
	return count
}

func (s *sessionRegistryImpl) removeSession(sessionID string) {
	s.registryLock.Lock()
	delete(s.sessionMap, sessionID)
	s.registryLock.Unlock()
}
