// Package retention defines validated, versioned retention configuration. It
// contains no application identities, routes, storage queries or authorization.
package retention

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ErrConflict = errors.New("retention policy version changed")

type Policy struct {
	RetentionDays          int `json:"retentionDays"`
	CleanupIntervalSeconds int `json:"cleanupIntervalSeconds"`
	CleanupBatchSize       int `json:"cleanupBatchSize"`
	CleanupMaxBatches      int `json:"cleanupMaxBatches"`
}

// UnmarshalJSON requires all four limits, including an explicit zero for forever.
func (p *Policy) UnmarshalJSON(data []byte) error {
	var raw struct {
		RetentionDays          *int `json:"retentionDays"`
		CleanupIntervalSeconds *int `json:"cleanupIntervalSeconds"`
		CleanupBatchSize       *int `json:"cleanupBatchSize"`
		CleanupMaxBatches      *int `json:"cleanupMaxBatches"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return err
	}
	if raw.RetentionDays == nil || raw.CleanupIntervalSeconds == nil || raw.CleanupBatchSize == nil || raw.CleanupMaxBatches == nil {
		return fmt.Errorf("all retention limits are required")
	}
	*p = Policy{*raw.RetentionDays, *raw.CleanupIntervalSeconds, *raw.CleanupBatchSize, *raw.CleanupMaxBatches}
	return p.Validate()
}

func (p Policy) Validate() error {
	if p.RetentionDays < 0 || p.RetentionDays > 36500 || p.CleanupIntervalSeconds < 1 || p.CleanupIntervalSeconds > 86400 || p.CleanupBatchSize < 1 || p.CleanupBatchSize > 5000 || p.CleanupMaxBatches < 1 || p.CleanupMaxBatches > 100 {
		return fmt.Errorf("invalid retention policy limits")
	}
	return nil
}

type Document struct {
	ResourceVersion int64             `json:"resourceVersion"`
	Policies        map[string]Policy `json:"policies"`
	UpdatedBy       string            `json:"updatedBy"`
	UpdateTime      int64             `json:"updateTime"`
}

func (d Document) Validate() error {
	if d.ResourceVersion < 0 || len(d.Policies) == 0 || len(d.Policies) > 16 {
		return fmt.Errorf("invalid retention document")
	}
	for key, p := range d.Policies {
		if key == "" || len(key) > 64 {
			return fmt.Errorf("invalid retention policy key")
		}
		if err := p.Validate(); err != nil {
			return err
		}
	}
	return nil
}
func Read(file string) (*Document, error) {
	info, err := os.Lstat(file)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 || info.Size() > 64<<10 {
		return nil, fmt.Errorf("retention policy file is not an owned regular file")
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var d Document
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&d); err != nil {
		return nil, err
	}
	if err = decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("invalid retention document payload")
	}
	if err = d.Validate(); err != nil {
		return nil, err
	}
	return &d, nil
}

// Save serializes writers across processes and atomically replaces an existing
// document. A missing or corrupt document is never silently overwritten.
func Save(file string, policies map[string]Policy, expected int64, actor string) (*Document, error) {
	if actor == "" || expected < 0 || expected == int64(^uint64(0)>>1) {
		return nil, fmt.Errorf("invalid retention update identity")
	}
	proposed := &Document{ResourceVersion: expected + 1, Policies: policies, UpdatedBy: actor, UpdateTime: time.Now().UnixMilli()}
	if err := proposed.Validate(); err != nil {
		return nil, err
	}
	unlock, err := lock(file + ".lock")
	if err != nil {
		return nil, err
	}
	defer unlock()
	current, err := Read(file)
	if err != nil {
		return nil, err
	}
	if current.ResourceVersion != expected {
		return nil, ErrConflict
	}
	data, err := json.Marshal(proposed)
	if err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(filepath.Dir(file), ".retention-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if _, err = tmp.Write(data); err == nil {
		err = tmp.Chmod(0644)
	}
	if err == nil {
		err = tmp.Sync()
	}
	closeErr := tmp.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	if err = os.Rename(tmp.Name(), file); err != nil {
		return nil, err
	}
	dir, err := os.Open(filepath.Dir(file))
	if err != nil {
		return nil, err
	}
	err = dir.Sync()
	closeErr = dir.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}
	return proposed, nil
}

// Controller polls a read-only policy projection, bounds concurrent cleanups,
// and uses the latest interval even when retention was disabled at startup.
type Controller struct {
	mu                sync.Mutex
	file, key         string
	fallback          Policy
	policy            Policy
	readErr           error
	nextRead, lastRun time.Time
	running           bool
}

func NewController(file, key string, fallback Policy, now time.Time) (*Controller, error) {
	if err := fallback.Validate(); err != nil {
		return nil, err
	}
	return &Controller{file: file, key: key, fallback: fallback, policy: fallback, lastRun: now}, nil
}
func (c *Controller) Begin(now time.Time) (Policy, bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	polled := !now.Before(c.nextRead)
	if polled {
		c.policy, c.readErr = c.fallback, nil
		if c.file != "" {
			document, err := Read(c.file)
			if err == nil {
				var ok bool
				c.policy, ok = document.Policies[c.key]
				if !ok {
					err = fmt.Errorf("required retention policy is missing")
				}
			} else if os.IsNotExist(err) {
				// Standalone services may use TOML without a projection mount. A mounted
				// directory missing its policy must stop cleanup, not expand retention.
				if _, directoryErr := os.Stat(filepath.Dir(c.file)); os.IsNotExist(directoryErr) {
					err = nil
				}
			}
			c.readErr = err
		}
		c.nextRead = now.Add(10 * time.Second)
	}
	if c.readErr != nil {
		if polled {
			return c.policy, false, c.readErr
		}
		return c.policy, false, nil
	}
	if c.running || c.policy.RetentionDays == 0 || now.Before(c.lastRun.Add(time.Duration(c.policy.CleanupIntervalSeconds)*time.Second)) {
		return c.policy, false, nil
	}
	c.running = true
	c.lastRun = now
	return c.policy, true, nil
}
func (c *Controller) Done() { c.mu.Lock(); c.running = false; c.mu.Unlock() }
