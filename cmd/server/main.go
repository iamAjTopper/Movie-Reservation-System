package main

import (
	"log"

	"movie-reservation/internal/db"
	"movie-reservation/internal/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading the .env file")
	}

	if err := db.Connect(); err != nil {
		log.Fatal("Database connection failed", err)
	}

	r := gin.Default()

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// Basic test routes
	r.GET("/me", routes.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": c.GetInt("user_id"),
			"role":    c.GetString("role"),
		})
	})

	r.GET("/admin/test",
		routes.AuthMiddleware(),
		routes.AdminOnlyMiddleware(),
		func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "welcome admin",
			})
		},
	)

	// 🔐 Auth
	routes.RegisterAuthRoutes(r)

	// 🎬 Movies
	routes.RegisterMovieRoutes(r)

	// 🕒 Showtimes
	routes.RegisterShowTimeRoutes(r)

	// 🎟 Seats
	routes.RegisterSeatRoutes(r)

	// 🎫 Reservations
	routes.RegisterReservationRoutes(r)

	// 👑 Admin
	routes.AdminReportsRoutes(r)
	routes.AdminUserRoutes(r)

	log.Printf("Server is running on :8080")
	r.Run(":8080")
}
