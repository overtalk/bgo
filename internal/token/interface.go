package itoken

import (
	"time"

	"github.com/overtalk/bgo/core"
)

const ModuleName = "internal.token"

type (
	ExpiredCB func(userId string, module ITokenModule)

	GameToken struct {
		Token       string
		AccessCount uint32
		Kicked      bool
	}
)

type ITokenModule interface {
	core.IModule

	Size() int64
	SetToken(userId string, token string, ttl int64) string
	GetToken(userId string) (*GameToken, error)
	Verify(userId, token string) (*GameToken, error)
	DelToken(userId string)
	// SetTokenExpired set the token expired without del the token
	SetTokenExpired(userId string)
	SetTokenTTL(userId string, ttl int64)
	KickOut(userId string) bool
	SetExpiredCallback(duration time.Duration, cb ExpiredCB)
}
