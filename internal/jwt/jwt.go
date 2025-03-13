package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userId string, jwtSecret string) (string, error) {
	claims := jwt.MapClaims{
		"id":  userId,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
