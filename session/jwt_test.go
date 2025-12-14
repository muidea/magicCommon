package session

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeecodeJWT(t *testing.T) {
	// bf1b29da3b034a82833cb59bf695be48
	os.Setenv("HMAC_SECRET", "bf1b29da3b034a82833cb59bf695be48")
	jwt := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJYLUF1dGhFbnRpdHkiOnsiaWQiOjk4MTAwNzM3NTgzOTQ4OSwibmFtZSI6ImFkbWluaXN0cmF0b3IiLCJlSUQiOjk5OTksImVUeXBlIjoiYWNjb3VudCIsInN0YXR1cyI6MSwibmFtZXNwYWNlIjoicGFuZWwifSwiWC1BdXRoTmFtZXNwYWNlIjoicGFuZWwiLCJfc2Vzc2lvbklEIjoiYXhtNlNTa3JOeEhuZzR6VzRPNWJOTjdRd0FmYUc2aXIiLCJpbm5lckV4cGlyZVRpbWUiOjE3NjU2ODE4NzQwNDQsImlubmVyU2Vzc2lvblN0YXJ0VGltZSI6MTc2NTY4MTI3NDA0NH0.EET76SpaOhjDKdyQsRXY1pit6AXTgNylF5H9Zw-yc2Q"
	sessionImpl := decodeJWT(jwt)
	assert.NotNil(t, sessionImpl)
	assert.Equal(t, "axm6SSkrNxHng4zW4O5bNN7QwAfaG6ir", sessionImpl.id)
}
