package middleware

import (
	"database/sql"

	"christ-api/pkg/database"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func AdminOnly(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(int64)
	if !ok {
		return response.Error(c, 401, "unauthorized", nil)
	}

	// Check if user has admin role
	var roleID sql.NullInt64
	query := `SELECT role_id FROM users WHERE id = $1`
	err := database.DB.QueryRow(query, userID).Scan(&roleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 401, "user not found", nil)
		}
		return response.Error(c, 500, "database error", nil)
	}

	if !roleID.Valid {
		return response.Error(c, 403, "no role assigned", nil)
	}

	// Check if role is admin
	var roleName string
	query = `SELECT name FROM roles WHERE id = $1`
	err = database.DB.QueryRow(query, roleID.Int64).Scan(&roleName)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 403, "role not found", nil)
		}
		return response.Error(c, 500, "database error", nil)
	}

	if roleName != "admin" {
		return response.Error(c, 403, "admin role required", nil)
	}

	return c.Next()
}
