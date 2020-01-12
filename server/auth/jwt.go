package auth

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

func getToken(id string) (string, error) {
	signingKey := []byte("keymaker")
	ttl := 30 * time.Second
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().UTC().Add(ttl).Unix(),
	})
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func verifyToken(tokenString string) (jwt.Claims, error) {
	signingKey := []byte("keymaker")
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}
