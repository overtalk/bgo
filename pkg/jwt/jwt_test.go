package jwtpkg_test

import (
	"testing"

	"github.com/overtalk/bgo/pkg/jwt"
)

func TestCJwt(t *testing.T) {
	jwt, err := jwtpkg.NewJwt([]byte("sdfsdfsdfs"), "HS256")
	if err != nil {
		t.Error(err)
		return
	}

	str, err := jwt.GenJwt(map[string]interface{}{
		"key1": "value1",
		"key2": 21,
	})

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(str)

	claims, err := jwt.Verify(str)
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("%+v\n", claims)
}
