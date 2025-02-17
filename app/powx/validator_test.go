package powx_test

import (
	"crypto/rand"
	"encoding/binary"
	"testing"
	"time"

	"pow-anti-ddos/app/powx"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/argon2"
)

func TestParseUnverifiedChallenge_Valid(t *testing.T) {
	secretKey := []byte("supersecretkey")
	params := powx.Params{Difficulty: 5, MemoryCost: 1024, TimeCost: 2, Threads: 4}
	challenge, err := powx.GenerateChallenge("testRequestID", secretKey, "127.0.0.1", params, time.Minute)
	assert.NoError(t, err)

	parsedParams, err := powx.ParseUnverifiedChallenge(challenge)
	assert.NoError(t, err)
	assert.Equal(t, params.Difficulty, parsedParams.Difficulty)
	assert.Equal(t, params.MemoryCost, parsedParams.MemoryCost)
	assert.Equal(t, params.TimeCost, parsedParams.TimeCost)
	assert.Equal(t, params.Threads, parsedParams.Threads)
}

func TestParseUnverifiedChallenge_InvalidToken(t *testing.T) {
	_, err := powx.ParseUnverifiedChallenge("invalid.token.here")
	assert.Error(t, err)
}

func TestParseAndVerifyChallenge_Valid(t *testing.T) {
	secretKey := []byte("supersecretkey")
	params := powx.Params{Difficulty: 5, MemoryCost: 1024, TimeCost: 2, Threads: 4}
	xff := "127.0.0.1"
	challenge, err := powx.GenerateChallenge("testRequestID", secretKey, xff, params, time.Minute)
	assert.NoError(t, err)

	parsedParams, err := powx.ParseAndVerifyChallenge(xff, challenge, secretKey)
	assert.NoError(t, err)
	assert.Equal(t, params.Difficulty, parsedParams.Difficulty)
	assert.Equal(t, params.MemoryCost, parsedParams.MemoryCost)
	assert.Equal(t, params.TimeCost, parsedParams.TimeCost)
	assert.Equal(t, params.Threads, parsedParams.Threads)
}

func TestParseAndVerifyChallenge_WrongXFF(t *testing.T) {
	secretKey := []byte("supersecretkey")
	params := powx.Params{Difficulty: 5, MemoryCost: 1024, TimeCost: 2, Threads: 4}
	challenge, err := powx.GenerateChallenge("testRequestID", secretKey, "127.0.0.1", params, time.Minute)
	assert.NoError(t, err)

	_, err = powx.ParseAndVerifyChallenge("192.168.0.1", challenge, secretKey)
	assert.Error(t, err)
}

func TestParseAndVerifyChallenge_InvalidSecret(t *testing.T) {
	secretKey := []byte("supersecretkey")
	wrongKey := []byte("wrongkey")
	params := powx.Params{Difficulty: 5, MemoryCost: 1024, TimeCost: 2, Threads: 4}
	challenge, err := powx.GenerateChallenge("testRequestID", secretKey, "127.0.0.1", params, time.Minute)
	assert.NoError(t, err)

	_, err = powx.ParseAndVerifyChallenge("127.0.0.1", challenge, wrongKey)
	assert.Error(t, err)
}

func TestIsValidSolution_Valid(t *testing.T) {
	params := powx.Params{
		Difficulty: 5,
		MemoryCost: 1024,
		TimeCost:   2,
		Threads:    4,
	}
	challenge := "test-challenge"
	solution := "correct-solution"

	for i := 0; i < 10000; i++ {
		attempt := make([]byte, 16)
		rand.Read(attempt)
		hashAttempt := argon2.IDKey(attempt, []byte(challenge), params.TimeCost, params.MemoryCost, params.Threads, 8)
		hashUint := binary.LittleEndian.Uint64(hashAttempt)
		if hashUint>>(64-params.Difficulty) == 0 {
			solution = string(attempt)
			break
		}
	}

	assert.True(t, powx.IsValidSolution(challenge, &params, solution))
}

func TestIsValidSolution_Invalid(t *testing.T) {
	params := powx.Params{
		Difficulty: 5,
		MemoryCost: 1024,
		TimeCost:   2,
		Threads:    4,
	}
	challenge := "test-challenge"
	solution := "incorrect-solution"

	assert.False(t, powx.IsValidSolution(challenge, &params, solution))
}
