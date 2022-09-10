package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/robrotheram/gogallery/backend/config"
)

var signingKey = []byte(config.RandomPassword(20))

func getToken(id string) (string, error) {
	ttl := 3000 * time.Second
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":  id,
		"exp": time.Now().UTC().Add(ttl).Unix(),
	})
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func VerifyToken(tokenString string) (jwt.Claims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}
