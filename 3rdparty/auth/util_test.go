package auth_test

import (
	"testing"

	"github.com/overtalk/bgo/3rdparty/auth"
)

func BenchmarkGenerateGameToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		auth.GenerateGameToken()
	}
}

func BenchmarkGenerateGameTokenSecret(b *testing.B) {
	for i := 0; i < b.N; i++ {
		auth.GenerateGameTokenSecret()
	}
}

func BenchmarkGenerateNonce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		auth.GenerateNonce()
	}
}
