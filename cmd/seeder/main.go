package main

import (
	"fmt"
	"log"

	"christ-api/internal/auth/helpers"
	"christ-api/pkg/database"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("ℹ️ .env.local/.env tidak ditemukan, pakai environment variables")
		}
	}

	database.Connect()
	defer database.DB.Close()

	// Create admin role if it doesn't exist
	var adminRoleID int64
	err := database.DB.QueryRow("SELECT id FROM roles WHERE name = 'admin' LIMIT 1").Scan(&adminRoleID)
	if err != nil {
		// Admin role doesn't exist, create it
		err = database.DB.QueryRow(
			"INSERT INTO roles (name, description) VALUES ('admin', 'Administrator role with full access') RETURNING id",
		).Scan(&adminRoleID)
		if err != nil {
			log.Fatalf("❌ Failed to create admin role: %v", err)
		}
		fmt.Printf("✅ Admin role created with ID: %d\n", adminRoleID)
	}

	email := "admin@christapi.dev"
	password := "admin123456"
	username := "admin"

	// Hash password
	hashedPassword, err := helpers.HashPassword(password)
	if err != nil {
		log.Fatalf("❌ Failed to hash password: %v", err)
	}

	// Check if admin already exists
	var exists bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		log.Fatalf("❌ Database error: %v", err)
	}

	if exists {
		log.Printf("⚠️ Admin user already exists with email: %s", email)
		return
	}

	// Insert admin user with admin role (already approved and active)
	query := `INSERT INTO users (email, username, password_hash, auth_provider, approval_status, is_active, role_id, created_at, updated_at)
	VALUES ($1, $2, $3, 'credentials', 'approved', TRUE, $4, NOW(), NOW())
	RETURNING id, email, username`

	var id int64
	var returnedEmail, returnedUsername string

	err = database.DB.QueryRow(query, email, username, hashedPassword, adminRoleID).Scan(&id, &returnedEmail, &returnedUsername)
	if err != nil {
		log.Fatalf("❌ Failed to create admin user: %v", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Printf("   ID: %d\n", id)
	fmt.Printf("   Email: %s\n", returnedEmail)
	fmt.Printf("   Username: %s\n", returnedUsername)
	fmt.Printf("   Password: %s\n", password)
	fmt.Printf("   Role ID: %d (admin)\n", adminRoleID)
	fmt.Println("\n📝 Note: Change this password after first login!")
}
