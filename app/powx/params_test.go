package powx_test

import (
	"testing"

	"pow-anti-ddos/app/powx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToClaims(t *testing.T) {
	params := powx.Params{
		Difficulty: 5,
		MemoryCost: 1024,
		TimeCost:   2,
		Threads:    4,
	}

	claims := params.ToClaims()

	assert.Equal(t, float64(params.Difficulty), claims["difficulty"])
	assert.Equal(t, float64(params.MemoryCost), claims["memory_cost"])
	assert.Equal(t, float64(params.TimeCost), claims["time_cost"])
	assert.Equal(t, float64(params.Threads), claims["threads"])
}

func TestNewParamsFromClaims(t *testing.T) {
	claims := jwt.MapClaims{
		"difficulty":  float64(5),
		"memory_cost": float64(1024),
		"time_cost":   float64(2),
		"threads":     float64(4),
	}

	params := powx.NewParamsFromClaims(claims)

	assert.Equal(t, 5, params.Difficulty)
	assert.Equal(t, uint32(1024), params.MemoryCost)
	assert.Equal(t, uint32(2), params.TimeCost)
	assert.Equal(t, uint8(4), params.Threads)
}

func TestNewParamsFromClaims_InvalidData(t *testing.T) {
	claims := jwt.MapClaims{
		"difficulty": "invalid",
	}

	require.Panics(t, func() {
		powx.NewParamsFromClaims(claims)
	})
}
