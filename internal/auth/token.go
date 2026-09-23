package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func GenerateToken(secret string, expireSeconds int64, userID int64, username string) (string, int64, error) {
	now := time.Now()
	expiresAt := now.Add(time.Duration(expireSeconds) * time.Second).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":   userID,
		"username": username,
		"iat":      now.Unix(),
		"exp":      expiresAt,
	})

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", 0, err
	}

	return signed, expiresAt, nil
}
