package session

import "time"

func NewAnonymousSession(remoteAddress, userAgent string) Session {
	now := time.Now().UTC()
	return &sessionImpl{
		id: createUUID(),
		context: map[string]any{
			InnerStartTime:        now.UnixMilli(),
			InnerRemoteAccessAddr: remoteAddress,
			InnerUseAgent:         userAgent,
			innerExpireTime:       now.Add(GetSessionTimeOutValue()).UnixMilli(),
		},
		observer: map[string]Observer{},
		status:   sessionActive,
	}
}
