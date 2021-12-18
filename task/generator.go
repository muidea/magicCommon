package task

import "github.com/muidea/magicCommon/foundation/generator"

func NewGeneratorTask(generator generator.Generator) *GeneratorTask {
	return &GeneratorTask{generator: generator}
}

type GeneratorTask struct {
	generator generator.Generator

	genCode string
}

func (s *GeneratorTask) Run() {
	s.genCode = s.generator.GenCode()
}

func (s *GeneratorTask) Result() string {
	return s.genCode
}
