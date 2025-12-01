package session

import (
	"errors"
)

func GetSessionValue[T any](session Session, key string) (ret T, err error) {
	rawVal, rawOK := session.GetOption(key)
	if !rawOK {
		err = errors.New("key not found")
		return
	}

	realVal, realOK := rawVal.(T)
	if !realOK {
		err = errors.New("type mismatch")
		return
	}

	ret = realVal
	return
}
