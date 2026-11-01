package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	cd "github.com/muidea/magicCommon/def"
)

type Status string

const (
	StatusStarting Status = "starting"
	StatusReady    Status = "ready"
	StatusFailed   Status = "failed"
	StatusUnready  Status = "unready"
	StatusStopping Status = "stopping"
)

type DependencyKind string

const (
	RequiredDependency DependencyKind = "required"
	OptionalDependency DependencyKind = "optional"
)

type Dependency struct {
	Name   string         `json:"-"`
	Kind   DependencyKind `json:"kind"`
	Target string         `json:"target"`
}

func Required(name string) Dependency {
	return Dependency{Name: name, Kind: RequiredDependency}
}

func Optional(name string) Dependency {
	return Dependency{Name: name, Kind: OptionalDependency}
}

type CheckStatus struct {
	Kind   DependencyKind `json:"kind"`
	Status string         `json:"status"`
	Target string         `json:"target,omitempty"`
	Error  string         `json:"error,omitempty"`
}

type ReadyResponse struct {
	Service   string                 `json:"service"`
	Status    Status                 `json:"status"`
	StartedAt string                 `json:"startedAt,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Checks    map[string]CheckStatus `json:"checks,omitempty"`
	CheckedAt string                 `json:"checkedAt,omitempty"`
}

type Manager struct {
	mu           sync.RWMutex
	probeMu      sync.Mutex
	dependencies map[string]Dependency
	checkedAt    time.Time
	maxAge       time.Duration
	service      string
	status       Status
	startedAt    time.Time
	errMsg       string
	checks       map[string]CheckStatus
	checkers     map[string]func(context.Context) *cd.Error
	client       *http.Client
}

var (
	defaultManager *Manager
	defaultOnce    sync.Once
)

func DefaultManager() *Manager {
	defaultOnce.Do(func() {
		defaultManager = NewManager()
	})
	return defaultManager
}

func ResetDefaultManager() {
	defaultManager = nil
	defaultOnce = sync.Once{}
}

func NewManager() *Manager {
	return &Manager{
		status:       StatusStarting,
		checks:       map[string]CheckStatus{},
		checkers:     map[string]func(context.Context) *cd.Error{},
		dependencies: map[string]Dependency{},
		client:       &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }},
	}
}

func (s *Manager) SetService(service string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.service = service
}

func (s *Manager) MarkStarting() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = StatusStarting
	s.errMsg = ""
	s.startedAt = time.Now().UTC()
	s.checks = map[string]CheckStatus{}
	s.checkedAt = time.Time{}
	s.maxAge = 0
}

func (s *Manager) RegisterDependencyChecker(name string, checker func(context.Context) *cd.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.checkers[name] = checker
}

func (s *Manager) MarkReady() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == StatusStopping {
		return
	}
	s.status = StatusReady
	s.errMsg = ""
}

// MarkStopping closes readiness before any input or dependency is released.
func (s *Manager) MarkStopping() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = StatusStopping
	s.errMsg = ""
}

func (s *Manager) MarkFailed(err *cd.Error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = StatusFailed
	if err != nil {
		s.errMsg = err.Error()
	}
}

func (s *Manager) Snapshot() ReadyResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	checks := map[string]CheckStatus{}
	for key, val := range s.checks {
		checks[key] = val
	}

	resp := ReadyResponse{
		Service: s.service,
		Status:  s.status,
		Checks:  checks,
	}
	if !s.startedAt.IsZero() {
		resp.StartedAt = s.startedAt.Format(time.RFC3339)
	}
	if s.errMsg != "" {
		resp.Error = s.errMsg
	}
	if !s.checkedAt.IsZero() {
		resp.CheckedAt = s.checkedAt.Format(time.RFC3339Nano)
	}
	hasRequired := false
	for _, check := range s.checks {
		if check.Kind == RequiredDependency {
			hasRequired = true
			break
		}
	}
	if resp.Status == StatusReady && s.maxAge > 0 && hasRequired {
		if s.checkedAt.IsZero() || time.Since(s.checkedAt) >= s.maxAge {
			resp.Status = StatusUnready
			resp.Error = "dependency checks are stale"
		} else {
			for _, check := range s.checks {
				if check.Kind == RequiredDependency && check.Status != "ready" {
					resp.Status = StatusUnready
					resp.Error = "required dependency is unavailable"
					break
				}
			}
		}
	}
	return resp
}

// RegisterDependency declares an infrastructure dependency owned by the
// registering adapter. It does not grant access or infer business policy.
func (s *Manager) RegisterDependency(dep Dependency) *cd.Error {
	if err := validateDependency(dep); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if prior, ok := s.dependencies[dep.Name]; ok && prior != dep {
		return cd.NewError(cd.IllegalParam, "conflicting dependency declaration")
	}
	s.dependencies[dep.Name] = dep
	return nil
}

// Dependencies freezes configured and adapter-owned declarations for one
// startup. A registered checker without a declaration is required by default.
func (s *Manager) Dependencies(configured []Dependency) ([]Dependency, *cd.Error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	merged := map[string]Dependency{}
	for name, dep := range s.dependencies {
		merged[name] = dep
	}
	for _, dep := range configured {
		if err := validateDependency(dep); err != nil {
			return nil, err
		}
		if prior, ok := merged[dep.Name]; ok && prior != dep {
			return nil, cd.NewError(cd.IllegalParam, "conflicting dependency declaration")
		}
		merged[dep.Name] = dep
	}
	for name := range s.checkers {
		if _, ok := merged[name]; !ok {
			merged[name] = Required(name)
		}
	}
	names := make([]string, 0, len(merged))
	for name := range merged {
		names = append(names, name)
	}
	sort.Strings(names)
	deps := make([]Dependency, 0, len(names))
	for _, name := range names {
		deps = append(deps, merged[name])
	}
	return deps, nil
}

func validateDependency(dep Dependency) *cd.Error {
	if strings.TrimSpace(dep.Name) == "" || dep.Name != strings.TrimSpace(dep.Name) || len(dep.Name) > 128 || (dep.Kind != RequiredDependency && dep.Kind != OptionalDependency) {
		return cd.NewError(cd.IllegalParam, "invalid dependency declaration")
	}
	if dep.Target != "" {
		target, err := url.Parse(dep.Target)
		if err != nil || (target.Scheme != "http" && target.Scheme != "https") || target.Hostname() == "" || target.User != nil || target.RawQuery != "" || target.Fragment != "" {
			return cd.NewError(cd.IllegalParam, "dependency target must be an HTTP origin without credentials, query or fragment")
		}
	}
	return nil
}

// CheckDependencies publishes an entire generation atomically. Readers keep
// the last completed generation while a probe is in flight, bounded by maxAge.
// Required checks run first; optional checks only use the remaining budget.
// Checkers must honor ctx; monitoring never launches detached timeout workers.
func (s *Manager) CheckDependencies(ctx context.Context, deps []Dependency) *cd.Error {
	if ctx == nil {
		return cd.NewError(cd.IllegalParam, "dependency check context is required")
	}
	if !s.probeMu.TryLock() {
		return cd.NewError(cd.ResourceExhausted, "dependency check is already running")
	}
	defer s.probeMu.Unlock()
	checks := map[string]CheckStatus{}
	var failure *cd.Error
	seen := make(map[string]struct{}, len(deps))
	ordered := make([]Dependency, 0, len(deps))
	for _, dep := range deps {
		if err := validateDependency(dep); err != nil {
			return err
		}
		if _, exists := seen[dep.Name]; exists {
			return cd.NewError(cd.IllegalParam, "duplicate dependency declaration")
		}
		seen[dep.Name] = struct{}{}
		if dep.Kind == RequiredDependency {
			ordered = append(ordered, dep)
		}
	}
	for _, dep := range deps {
		if dep.Kind == OptionalDependency {
			ordered = append(ordered, dep)
		}
	}
	for _, dep := range ordered {
		check := CheckStatus{Kind: dep.Kind, Target: dep.Target, Status: "ready"}
		if err := s.checkDependency(ctx, dep); err != nil {
			// Never expose driver errors, credentials or response bodies on /ready.
			check.Status = "failed"
			check.Error = "dependency check failed"
			if dep.Kind == RequiredDependency {
				failure = cd.NewError(cd.ServiceUnavailable, "required dependency is unavailable")
			}
		}
		checks[dep.Name] = check
	}
	s.mu.Lock()
	s.checks = checks
	s.checkedAt = time.Now().UTC()
	s.mu.Unlock()
	return failure
}

func (s *Manager) checkDependency(ctx context.Context, dep Dependency) (err *cd.Error) {
	defer func() {
		if recover() != nil {
			err = cd.NewError(cd.ServiceUnavailable, "dependency checker panicked")
		}
	}()
	if ctx.Err() != nil {
		return cd.NewError(cd.Timeout, "dependency check canceled")
	}
	s.mu.RLock()
	checker := s.checkers[dep.Name]
	s.mu.RUnlock()
	if checker != nil {
		err = checker(ctx)
		if ctx.Err() != nil {
			return cd.NewError(cd.Timeout, "dependency check canceled")
		}
		return err
	}
	if dep.Target == "" {
		return cd.NewError(cd.ServiceUnavailable, "dependency has no checker or target")
	}
	req, requestErr := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(dep.Target, "/")+"/health/ready", nil)
	if requestErr != nil {
		return cd.NewError(cd.IllegalParam, "invalid dependency request")
	}
	resp, requestErr := s.client.Do(req)
	if requestErr != nil {
		return cd.NewError(cd.ServiceUnavailable, "dependency request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return cd.NewError(cd.ServiceUnavailable, fmt.Sprintf("dependency not ready, status:%d", resp.StatusCode))
	}
	if ctx.Err() != nil {
		return cd.NewError(cd.Timeout, "dependency check canceled")
	}
	return nil
}

func (s *Manager) LiveHandler(_ context.Context, res http.ResponseWriter, _ *http.Request) {
	writeJSON(res, http.StatusOK, map[string]string{"status": "live"})
}

func (s *Manager) ReadyHandler(_ context.Context, res http.ResponseWriter, _ *http.Request) {
	snapshot := s.Snapshot()
	statusCode := http.StatusOK
	if snapshot.Status != StatusReady {
		statusCode = http.StatusServiceUnavailable
	}
	writeJSON(res, statusCode, snapshot)
}

func writeJSON(res http.ResponseWriter, statusCode int, value any) {
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(statusCode)
	_ = json.NewEncoder(res).Encode(value)
}

func (s DependencyKind) String() string {
	return string(s)
}
