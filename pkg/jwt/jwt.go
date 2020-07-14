package jwtpkg

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

type Jwt struct {
	secret     []byte
	signMethod jwt.SigningMethod
}

func NewJwt(secret []byte, alg string) (*Jwt, error) {
	signMethod := jwt.GetSigningMethod(alg)
	if signMethod == nil {
		return nil, errors.New("invalid sign alg : " + alg)
	}

	return &Jwt{
		secret:     secret,
		signMethod: signMethod,
	}, nil
}

func (this *Jwt) GenJwt(claims jwt.MapClaims) (string, error) {
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(this.signMethod, claims)
	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString(this.secret)
}

func (this *Jwt) parseFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}
	// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
	return this.secret, nil
}

func (this *Jwt) Verify(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, this.parseFunc)
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}
