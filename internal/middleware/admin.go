package middleware

import (
	"strings"

	jwtpkg "christ-api/pkg/jwt"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func AdminOnly(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return response.Error(c, 401, "missing token", nil)
	}

	tokenString := strings.Split(authHeader, " ")
	if len(tokenString) != 2 {
		return response.Error(c, 401, "invalid token format", nil)
	}

	token, err := jwtlib.Parse(tokenString[1], func(t *jwtlib.Token) (interface{}, error) {
		return jwtpkg.Secret(), nil
	})

	if err != nil || !token.Valid {
		return response.Error(c, 401, "invalid token", nil)
	}

	var roleID int64
	if claims, ok := token.Claims.(jwtlib.MapClaims); ok {
		if rid, exists := claims["role_id"]; exists {
			switch v := rid.(type) {
			case float64:
				roleID = int64(v)
			case int64:
				roleID = v
			}
		}
	}

	if roleID == 0 {
		return response.Error(c, 403, "no role assigned", nil)
	}

	if roleID != 1 {
		return response.Error(c, 403, "admin role required", nil)
	}

	return c.Next()
}

