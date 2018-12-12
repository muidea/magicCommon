package builder

import "muidea.com/magicCommon/orm/mysql"

// Builder orm builder
type Builder interface {
	BuildSchema() (string, error)
}

// NewBuilder new builder
func NewBuilder(obj interface{}) Builder {
	return mysql.New(obj)
}
