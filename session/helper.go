package session

func GetSessionValue[T any](session Session, key string) (ret T, ok bool) {
	rawVal, rawOK := session.GetOption(key)
	if !rawOK {
		ok = rawOK
		return
	}

	realVal, realOK := rawVal.(T)
	if !realOK {
		ok = realOK
		return
	}

	ret = realVal
	return
}
