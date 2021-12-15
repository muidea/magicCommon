package module

type Module interface {
	ID() string
	Startup()
	Shutdown()
}

type Service interface {
	Name() string
	Startup()
	Run()
	Shutdown()
}

var name2ModuleList map[string][]Module

func Register(serviceName string, module Module) {
	list, ok := name2ModuleList[serviceName]
	if !ok {
		list = []Module{}
	}

	list = append(list, module)
	name2ModuleList[serviceName] = list
}

func GetModules(serviceName string) []Module {
	list, ok := name2ModuleList[serviceName]
	if ok {
		return list
	}

	return []Module{}
}
