package powx_test

import (
	"testing"
	"time"

	"pow-anti-ddos/app/powx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateChallenge(t *testing.T) {
	secretKey := []byte("mysecretkey")
	requestID := "test-request"
	xff := "192.168.1.1"
	params := powx.Params{
		Difficulty: 1,
		MemoryCost: 2,
		TimeCost:   3,
		Threads:    4,
	}
	ttl := 5 * time.Minute

	challenge, err := powx.GenerateChallenge(requestID, secretKey, xff, params, ttl)
	if err != nil {
		t.Fatalf("failed to generate challenge: %v", err)
	}

	parsedToken, err := jwt.Parse(challenge, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})
	require.NoError(t, err)
	require.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, xff, claims["sub"])
	assert.Equal(t, requestID, claims["jti"])
	assert.EqualValues(t, 1, claims["difficulty"])
	assert.EqualValues(t, 2, claims["memory_cost"])
	assert.EqualValues(t, 3, claims["time_cost"])
	assert.EqualValues(t, 4, claims["threads"])
	now := time.Now().Unix()
	iat, ok := claims["iat"].(float64)
	if !ok || int64(iat) > now {
		t.Errorf("invalid iat: %v", iat)
	}
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) <= now {
		t.Errorf("invalid exp: %v", exp)
	}
	assert.EqualValues(t, exp, iat+ttl.Seconds())
}
