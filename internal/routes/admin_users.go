package routes

import (
	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
)

func AdminUserRoutes(r *gin.Engine) {
	r.PUT("/admin/users/:id/promote", AuthMiddleware(), AdminOnlyMiddleware(), promoteUser)
}

func promoteUser(c *gin.Context) {
	userId := c.Param("id")

	result, err := db.DB.Exec(`
		UPDATE users
		SET role = 'ADMIN'
		WHERE id = $1
		`, userId)

	if err != nil {
		c.JSON(500, gin.H{"error": "failed to promote user"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}
	c.JSON(200, gin.H{"message": "user promoted to admin"})
}
