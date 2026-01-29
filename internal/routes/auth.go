package routes

import "github.com/gin-gonic/gin"

func RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/auth/register", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implemented"})
	})

	r.POST("/auth/login", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implemented"})
	})
}
