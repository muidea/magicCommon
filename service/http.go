package service

import (
	"github.com/muidea/magicCommon/event"
	"github.com/muidea/magicCommon/module"
	"github.com/muidea/magicCommon/task"
	engine "github.com/muidea/magicEngine"
)

// NewHTTP 新建Http Service
func NewHTTP(endpointName, listenPort string) (ret Service, err error) {
	core := &http{
		endpointName: endpointName,
		listenPort:   listenPort,
	}

	ret = core
	return
}

// http Core对象
type http struct {
	endpointName string
	listenPort   string

	httpServer engine.HTTPServer
}

// Startup 启动
func (s *http) Startup(
	eventHub event.Hub,
	backgroundRoutine task.BackgroundRoutine) {
	router := engine.NewRouter()

	s.httpServer = engine.NewHTTPServer(s.listenPort)
	s.httpServer.Bind(router)

	modules := module.GetModules()
	for _, val := range modules {
		val.Setup(s.endpointName, eventHub, backgroundRoutine)
	}
}

func (s *http) Run() {
	s.httpServer.Run()
}

// Shutdown 销毁
func (s *http) Shutdown() {
	modules := module.GetModules()
	for _, val := range modules {
		val.Teardown()
	}
}
