package util

type Filter interface {
	Enable(val interface{}) bool
}
