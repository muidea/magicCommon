package session

import (
	"fmt"

	"github.com/golang-jwt/jwt/v4"

	"github.com/muidea/magicCommon/foundation/log"
)

func SignatureJWT(mc jwt.MapClaims) (Token, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, mc)
	valStr, valErr := token.SignedString([]byte(getSecret()))
	if valErr != nil {
		log.Errorf("Signature failed, err:%s", valErr.Error())
		return "", valErr
	}

	return Token(valStr), nil
}

func decodeJWT(sigVal string) *sessionImpl {
	secretVal := getSecret()
	token, err := jwt.Parse(sigVal, func(token *jwt.Token) (any, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v ", token.Header["alg"])
		}

		// hmacSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secretVal), nil
	})
	if err != nil {
		//log.Infof("illegal jwt value:%s, secret:%s, err:%s", sigVal[1], secretVal, err.Error())
		return nil
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sessionPtr := &sessionImpl{context: map[string]any{}, observer: map[string]Observer{}}
		sessionPtr.context[InnerAuthType] = AuthJWTSession
		for k, v := range claims {
			if k == innerSessionID {
				sessionPtr.id = v.(string)
				continue
			}
			sessionPtr.context[k] = v
		}

		return sessionPtr
	}

	return nil
}
