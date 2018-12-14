package orm

import (
	"muidea.com/magicCommon/orm/model"
)

type serverConfig struct {
	user     string
	password string
	address  string
	dbName   string
}

type manager struct {
	serverConfig *serverConfig

	moduleInfoCache model.StructInfoCache
}

func newManager() *manager {
	return &manager{moduleInfoCache: model.NewCache()}
}

func (s *manager) updateServerConfig(cfg *serverConfig) {
	s.serverConfig = cfg
}

func (s *manager) getServerConfig() *serverConfig {
	return s.serverConfig
}

func (s *manager) getCache() model.StructInfoCache {
	return s.moduleInfoCache
}
