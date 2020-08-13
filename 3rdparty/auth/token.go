package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// GameToken game token
type GameToken struct {
	devID       string
	token       string
	tokenSecret string
	accessCount uint32
}

// NewGameToken create a *GameToken struct
func NewGameToken(devID string) *GameToken {
	return &GameToken{
		devID:       devID,
		token:       GenerateGameToken(),
		tokenSecret: GenerateGameTokenSecret(),
		accessCount: 0,
	}
}

var (
	// ErrTokenNotExist token isn't existed
	ErrTokenNotExist = errors.New("token not exist")
	// ErrTokenVerifyTooQuick too quick to verify the token
	ErrTokenVerifyTooQuick = errors.New("token verify too quick")
)

const (
	// maxGameTokenTTL define the expired token second
	maxGameTokenTTL = 60 * 60

	// minGameTokenAccessTime defines the time duration of two token request
	minGameTokenAccessTime = 60 * time.Millisecond
)

// GameTokenCache cache for game token
type GameTokenCache struct {
	tokens *Cache
}

// NewGameTokenCache create a *GameTokenCache struct
func NewGameTokenCache() *GameTokenCache {
	return &GameTokenCache{tokens: NewCache()}
}

func (this *GameTokenCache) Size() int64 {
	if this != nil && this.tokens != nil {
		return this.Size()
	}

	return 0
}

func (this *GameTokenCache) DelToken(userID string) {
	// set it to be expired ASAP
	this.tokens.Delete(userID)
}

const kickoutPrefix = "kickout:"

// Kickout kickout a user by userid
// just change the token
func (this *GameTokenCache) KickOut(userID string) bool {
	// no need to update token's access time
	item, _, flag := this.tokens.Get(userID, false)
	if !flag || item == nil {
		return false
	}

	t := item.(*GameToken)
	t.token = fmt.Sprintf("%s%d", kickoutPrefix, time.Now().UnixNano())
	return true
}

// GetToken return ( devID, token, tokenSecret )
func (this *GameTokenCache) GetToken(userID string) (string, string, string) {
	// no need to update token's access time
	item, _, ok := this.tokens.Get(userID, false)
	if !ok || item == nil {
		return "", "", ""
	}

	t := item.(*GameToken)
	if strings.HasPrefix(t.token, kickoutPrefix) {
		return "", "", ""
	}
	return t.devID, t.token, t.tokenSecret
}

func (this *GameTokenCache) SetToken(userID, devID string) (string, string) {
	ttl := int64(maxGameTokenTTL)
	hour, min, sec := time.Now().Clock()
	if hour == 4 {
		// at 00:05:00, token will be exipred, to reset user data quickly
		ttl = maxGameTokenTTL - int64(min*60+sec)
	}

	tk := NewGameToken(devID)
	this.tokens.Set(userID, tk, ttl)
	return tk.token, tk.tokenSecret
}

func (this *GameTokenCache) Verify(userID string, token string) (string, error) {
	item, deltaTime, ok := this.tokens.Get(userID, true)
	if !ok || item == nil {
		return "", ErrTokenNotExist
	}

	t := item.(*GameToken)
	if token != t.token {
		return "", fmt.Errorf("bad game_token, %s != %s", token, t.token)
	}

	t.accessCount++
	if t.accessCount > 5 && deltaTime < minGameTokenAccessTime {
		// TODO: add log
		//zap.S().Errorf("UserID: %s, TokenDeltaTime: %dms",
		//	userID, deltaTime/time.Millisecond,
		//)
		return "", ErrTokenVerifyTooQuick
	}

	return t.tokenSecret, nil
}
