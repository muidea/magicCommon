package orm

import (
	"fmt"
	"sync"
)

type serverConfig struct {
	user     string
	password string
	address  string
	dbName   string
}

type manager struct {
	serverConfig *serverConfig

	// name->pkgPath
	modelInfo map[string]string
	modelLock sync.RWMutex
}

func newManager() *manager {
	return &manager{modelInfo: map[string]string{}}
}

func (s *manager) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *manager) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *manager) registerModule(name, pkgPath string) error {
	s.modelLock.Lock()
	defer s.modelLock.Unlock()

	path, ok := s.modelInfo[name]
	if ok {
		if path != pkgPath {
			return fmt.Errorf("duplicate module, name:%s, existPath:%s, newPath:%s", name, path, pkgPath)
		}
	}
	s.modelInfo[name] = pkgPath

	return nil
}

func (s *manager) unregisterModule(name string) error {
	s.modelLock.Lock()
	defer s.modelLock.Unlock()

	_, ok := s.modelInfo[name]
	if !ok {
		return fmt.Errorf("no found module, name:%s", name)
	}

	delete(s.modelInfo, name)
	return nil
}

func (s *manager) findModule(name string) (string, error) {
	s.modelLock.RLock()
	defer s.modelLock.RUnlock()

	path, ok := s.modelInfo[name]
	if ok {
		return path, nil
	}

	return "", fmt.Errorf("no found module, name:%s", name)
}
