package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"io"
)

func makeRandomBytes(size int) []byte {
	b := make([]byte, size)
	rand.Read(b)
	return b
}

func makeRandomString() string {
	return hex.EncodeToString(makeRandomBytes(16))
}

// GenerateNonce generate a nonce string
func GenerateNonce() string {
	return hex.EncodeToString(makeRandomBytes(20))
}

// GenerateGameToken generate the game token
func GenerateGameToken() string {
	h := sha1.New()
	io.WriteString(h, makeRandomString()+"game_token_salt")
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateGameTokenSecret generate the game token seceret
func GenerateGameTokenSecret() string {
	h := sha1.New()
	io.WriteString(h, makeRandomString()+"game_token_secret_salt")
	return hex.EncodeToString(h.Sum(nil))
}
