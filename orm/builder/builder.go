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
	GetTableName() string

	GetRelationTableName(relationInfo model.StructInfo) string
	BuildCreateRelationSchema(relationInfo model.StructInfo) (string, error)
	BuildDropRelationSchema(relationInfo model.StructInfo) (string, error)
	BuildInsertRelation(relationInfo model.StructInfo) (string, error)
	BuildUpdateRelation(relationInfo model.StructInfo) (string, error)
	BuildDeleteRelation(relationInfo model.StructInfo) (string, error)
	BuildQueryRelation(relationInfo model.StructInfo) (string, error)
}

// NewBuilder new builder
func NewBuilder(structInfo model.StructInfo) Builder {
	return mysql.New(structInfo)
}
