package builder

import (
	"muidea.com/magicCommon/orm/model"
	"muidea.com/magicCommon/orm/mysql"
)

// Builder orm builder
type Builder interface {
	BuildCreateSchema() (string, error)
	BuildDropSchema() (string, error)
	BuildInsert() (string, error)
	BuildUpdate() (string, error)
	BuildDelete() (string, error)
	BuildQuery() (string, error)
}

// NewBuilder new builder
func NewBuilder(structInfo *model.StructInfo) Builder {
	return mysql.New(structInfo)
}
