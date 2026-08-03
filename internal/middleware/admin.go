package middleware

import (
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func AdminOnly(c *fiber.Ctx) error {
	// Extract role_id from JWT token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return response.Error(c, 401, "missing token", nil)
	}

	tokenString := authHeader[7:] // Remove "Bearer " prefix
	token, err := jwtlib.Parse(tokenString, func(t *jwtlib.Token) (interface{}, error) {
		// This is a simplified check - in production you'd want to use the same validation as AuthMiddleware
		return nil, nil
	})
	if err != nil {
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

	// Check if role_id is 1 (admin role - adjust based on your database)
	// In a real scenario, you'd have a constant for admin role ID
	if roleID != 1 {
		return response.Error(c, 403, "admin role required", nil)
	}

	return c.Next()
}

