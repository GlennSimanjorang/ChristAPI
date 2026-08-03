package jwt

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secret []byte

func loadSecret() []byte {
	if len(secret) > 0 {
		return secret
	}

	if s := os.Getenv("JWT_SECRET"); s != "" {
		secret = []byte(s)
	}

	return secret
}

func Secret() []byte {
	return loadSecret()
}

func GenerateToken(userID int, roleID *int64) (string, error) {
	if len(loadSecret()) == 0 {
		return "", errors.New("JWT_SECRET is not configured")
	}

	now := time.Now().Unix()
	claims := jwt.MapClaims{
		"user_id": userID,
		"iat":     now,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	if roleID != nil {
		claims["role_id"] = *roleID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(loadSecret())
}
