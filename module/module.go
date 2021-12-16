package module

type Module interface {
	ID() string
	Setup()
	Teardown()
}

type Service interface {
	Name() string
	Startup()
	Run()
	Shutdown()
}

var moduleList []Module

func Register(module Module) {
	moduleList = append(moduleList, module)
}

func GetModules() []Module {
	return moduleList
}
