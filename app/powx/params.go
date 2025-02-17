package powx

import "github.com/golang-jwt/jwt/v5"

const (
	claimKeyDifficulty = "difficulty"
	claimKeyMemoryCost = "memory_cost"
	claimKeyTimeCost   = "time_cost"
	claimKeyThreads    = "threads"
)

type Params struct {
	Difficulty int
	MemoryCost uint32
	TimeCost   uint32
	Threads    uint8
}

func (p *Params) ToClaims() jwt.MapClaims {
	return jwt.MapClaims{
		claimKeyDifficulty: float64(p.Difficulty),
		claimKeyMemoryCost: float64(p.MemoryCost),
		claimKeyTimeCost:   float64(p.TimeCost),
		claimKeyThreads:    float64(p.Threads),
	}
}

func NewParamsFromClaims(claims jwt.MapClaims) *Params {
	return &Params{
		Difficulty: int(claims[claimKeyDifficulty].(float64)),
		MemoryCost: uint32(claims[claimKeyMemoryCost].(float64)),
		TimeCost:   uint32(claims[claimKeyTimeCost].(float64)),
		Threads:    uint8(claims[claimKeyThreads].(float64)),
	}
}
