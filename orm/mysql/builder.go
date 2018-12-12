package mysql

type builder struct {
	obj interface{}
}

func (s *builder) BuildSchema() string {
	modelInfo := orm.getModelInfo(s.obj)

}
