package main

import (
	"log"

	"movie-reservation/internal/db"

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

	log.Printf("Server is running on :8080")
	r.Run(":8080")
}
