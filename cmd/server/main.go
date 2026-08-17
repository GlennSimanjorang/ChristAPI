package main

import (
	"log"
	"os"

	"christ-api/internal/auth"
	"christ-api/internal/middleware"
	"christ-api/internal/role"
	"christ-api/pkg/database"
	"christ-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// load env dulu
	if err := godotenv.Load(".env.local"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("ℹ️ .env.local/.env tidak ditemukan, pakai environment variables")
		}
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("❌ JWT_SECRET wajib diisi")
	}

	// connect database
	database.Connect()

	// initialize services that need DB
	auth.InitService(&auth.AuthRepository{DB: database.DB})
	role.InitService(&role.RoleRepository{DB: database.DB})

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))
	app.Use(middleware.CustomLogger)

	// Serve static files from docs directory
	app.Static("/docs", "./docs")

	routes.Setup(app)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on :" + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
