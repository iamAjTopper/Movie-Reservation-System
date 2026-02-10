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

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	r.GET("/me", routes.AuthMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"user_id": c.GetInt("user_id"),
			"role":    c.GetString("role"),
		})
	})

	r.GET("/admin/test", routes.AuthMiddleware(), routes.AdminOnlyMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "welcome admin",
		})
	})

	log.Printf("Server is running on :8080")
	routes.RegisterAuthRoutes(r)
	r.Run(":8080")
}
