package powx

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateChallenge(requestID string, secretKey []byte, xff string, params Params, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := params.ToClaims()
	claims["sub"] = xff
	claims["iat"] = now.Unix()
	claims["exp"] = now.Add(ttl).Unix()
	claims["jti"] = requestID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(secretKey)
}
