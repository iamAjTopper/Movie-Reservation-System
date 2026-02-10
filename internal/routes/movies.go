package routes

import "github.com/gin-gonic/gin"

func RegisterMovieRoutes(r *gin.Engine) {
	//PUBLIC

	r.GET("/movies", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implemented"})
	})

	//admin

	r.POST("/movies", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implemenetd"})
	})
}
