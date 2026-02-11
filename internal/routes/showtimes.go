package routes

import (
	"github.com/gin-gonic/gin"
)

func RegisterShowTimeRoutes(r *gin.Engine) {
	//public
	r.GET("/movies/:id/showtimes", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implememted"})
	})

	//admin
	r.POST("/movies/:id/showtimes", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not imeplemented"})
	})
}
