package session

const (
	// sessionID 会话ID
	sessionID = "sessionID"
	// AuthNamespace Account namespace
	AuthNamespace = "sessionNamespace"
	// AuthEntity Entity
	AuthEntity = "authEntity"
	// AuthRole 权限组
	AuthRole = "authRole"
)

// AuthRemoteAddress 远端地址
const AuthRemoteAddress = "$$sessionRemoteAddress"

// AuthExpiryValue 会话有效期
const AuthExpiryValue = "$$sessionExpiryValue"

// refreshTime 会话刷新时间
const refreshTime = "$$sessionRefreshTime"
