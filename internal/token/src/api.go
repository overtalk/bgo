package ctoken

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/overtalk/bgo/internal/token"
	"github.com/overtalk/bgo/pkg/log"
)

const (
	maxGameTokenTTL        = 3600
	minGameTokenAccessTime = 60 * time.Millisecond
)

var (
	// ErrTokenNotExist token isn't existed
	ErrTokenNotExist = errors.New("token not exist")
	// ErrKickedPlayer player is kicked out
	ErrKickedPlayer = errors.New("player is kicked out")
	// ErrTokenVerifyTooQuick too quick to verify the token
	ErrTokenVerifyTooQuick = errors.New("token verify too quick")
)

// Size get the underlying tokens' size
func (tm *CTokenModule) Size() int64 { return tm.cache.Size() }

func (tm *CTokenModule) SetToken(userId string, token string, ttl int64) string {
	hour, min, sec := time.Now().Clock()
	if hour == 4 {
		// at 00:05:00, token will be exipred, to reset user data quickly
		ttl = maxGameTokenTTL - int64(min*60+sec)
	}
	tk := &itoken.GameToken{
		Token:       token,
		AccessCount: 0,
		Kicked:      false,
	}
	tm.cache.Set(userId, tk, ttl)
	return tk.Token
}

// GetToken get a game token by userId
func (tm *CTokenModule) GetToken(userId string) (*itoken.GameToken, error) {
	// no need to update token's access time
	item, _, ok := tm.cache.Get(userId, false)
	if !ok || item == nil {
		return nil, ErrTokenNotExist
	}
	t := item.(*itoken.GameToken)
	if t.Kicked {
		return nil, ErrKickedPlayer
	}
	return t, nil
}

func (tm *CTokenModule) Verify(userId, token string) (*itoken.GameToken, error) {
	item, deltaTime, ok := tm.cache.Get(userId, true)
	if !ok || item == nil {
		return nil, ErrTokenNotExist
	}
	t := item.(*itoken.GameToken)
	if token != t.Token {
		return nil, fmt.Errorf("bad game_token, %s != %s", token, t.Token)
	}
	t.AccessCount++
	if t.AccessCount > 5 && deltaTime < minGameTokenAccessTime {
		//logpkg.GetLogger().With(
		//	zap.String("user_id", userId),
		//	zap.Any("TokenDeltaTime", deltaTime/time.Millisecond),
		//).Error("verify token too quick")
		logpkg.Error("verify token too quick", zap.String("user_id", userId), zap.Any("TokenDeltaTime", deltaTime/time.Millisecond))

		return nil, ErrTokenVerifyTooQuick
	}
	return t, nil
}

// DelToken remove a game token by userId
func (tm *CTokenModule) DelToken(userId string) { tm.cache.Delete(userId) }

// SetTokenExpired set the token expired without del the token
func (tm *CTokenModule) SetTokenExpired(userId string) { tm.SetTokenTTL(userId, 0) }

func (tm *CTokenModule) SetTokenTTL(userId string, ttl int64) { tm.cache.SetExpiration(userId, ttl) }

// KickOut kick out a user by userId
func (tm *CTokenModule) KickOut(userId string) bool {
	// no need to update token's access time
	item, _, ok := tm.cache.Get(userId, false)
	if !ok || item == nil {
		return false
	}
	t := item.(*itoken.GameToken)
	t.Kicked = true
	return true
}

func (tm *CTokenModule) SetExpiredCallback(duration time.Duration, cb itoken.ExpiredCB) {
	tm.expiredDuration = duration
	tm.expiredCallback = cb
}
