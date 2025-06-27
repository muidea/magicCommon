package util

import (
	"github.com/google/uuid"
	"strings"
)

func NewUUID() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
