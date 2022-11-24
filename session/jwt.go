package session

import (
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	"github.com/golang-jwt/jwt/v4"
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
	token, err := jwt.Parse(sigVal, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v ", token.Header["alg"])
		}

		// hmacSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(getSecret()), nil
	})
	if err != nil {
		log.Infof("illegal jwt value:%s, err:%s", sigVal[1], err.Error())
		return nil
	}

	currentTime := time.Now().UTC()
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sessionPtr := &sessionImpl{context: map[string]interface{}{}, observer: map[string]Observer{}}
		for k, v := range claims {
			if k == sessionID {
				sessionPtr.id = v.(string)
				continue
			}

			if k == expiryTime {
				if v.(float64) < float64(currentTime.Unix()) {
					log.Infof("illegal jwt,expiry time")
					return nil
				}

				sessionPtr.context[k] = currentTime.Add(DefaultSessionTimeOutValue).Unix()
				continue
			}

			sessionPtr.context[k] = v
		}

		return sessionPtr
	}

	return nil
}
