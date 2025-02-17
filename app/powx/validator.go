package powx

import (
	"encoding/binary"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

func ParseUnverifiedChallenge(challenge string) (*Params, error) {
	token, _, err := jwt.NewParser().ParseUnverified(challenge, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return NewParamsFromClaims(claims), nil
	}
	return nil, fmt.Errorf("invalid token")
}

func ParseAndVerifyChallenge(xff, challenge string, secretKey []byte) (*Params, error) {
	parser := jwt.NewParser(jwt.WithSubject(xff), jwt.WithExpirationRequired())
	token, err := parser.Parse(challenge, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return NewParamsFromClaims(claims), nil
	}
	return nil, fmt.Errorf("invalid token")
}

func IsValidSolution(challenge string, params *Params, solution string) bool {
	hashBytes := argon2.IDKey([]byte(solution), []byte(challenge), params.TimeCost, params.MemoryCost, params.Threads, 8)
	hashUint64 := binary.LittleEndian.Uint64(hashBytes)
	return hashUint64>>(64-params.Difficulty) == 0
}
