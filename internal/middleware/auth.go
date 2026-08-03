package middleware

import (
	"database/sql"
	"strings"

	"christ-api/pkg/database"
	jwtpkg "christ-api/pkg/jwt"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return response.Error(c, 401, "missing token", nil)
	}

	// format: Bearer TOKEN
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

	var userID int64
	var tokenIssuedAt int64

	// try to extract user_id and iat claim and set to locals for handlers
	if claims, ok := token.Claims.(jwtlib.MapClaims); ok {
		if uid, exists := claims["user_id"]; exists {
			switch v := uid.(type) {
			case float64:
				userID = int64(v)
			case int64:
				userID = v
			case int:
				userID = int64(v)
			}
		}
		if iat, exists := claims["iat"]; exists {
			switch v := iat.(type) {
			case float64:
				tokenIssuedAt = int64(v)
			case int64:
				tokenIssuedAt = v
			case int:
				tokenIssuedAt = int64(v)
			}
		}
	}

	if userID == 0 {
		return response.Error(c, 401, "invalid token claims", nil)
	}

	// Check user active status, approval status, and last logout time in database
	var isActive bool
	var approvalStatus string
	var lastLogoutAt sql.NullTime
	query := `SELECT is_active, approval_status, COALESCE(last_logout_at, NULL) FROM users WHERE id = $1 LIMIT 1`
	err = database.DB.QueryRow(query, userID).Scan(&isActive, &approvalStatus, &lastLogoutAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 401, "user not found", nil)
		}
		return response.Error(c, 500, "database error", nil)
	}

	if !isActive || approvalStatus != "approved" {
		return response.Error(c, 403, "your account is inactive or pending approval", nil)
	}

	// Validate token wasn't issued before user's last logout
	if lastLogoutAt.Valid && tokenIssuedAt < lastLogoutAt.Time.Unix() {
		return response.Error(c, 401, "token has been invalidated by logout", nil)
	}

	c.Locals("user_id", userID)
	return c.Next()
}
