package routes

import (
	"net/http"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterAuthRoutes(r *gin.Engine) {
	r.POST("/auth/register", func(c *gin.Context) {
		//Binding: Move data from the Request Body into our Struct
		var req registerRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// 3. Validation: Check if any fields are empty
		if req.Email == "" || req.Password == "" || req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing fields"})
			return
		}

		//4. Security: NEVER store raw passwords.
		// bcrypt transforms "password123" into a long, unreadable string.
		// DefaultCost is the amount of work the CPU does to scramble it.
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "password hashing failed"})
			return
		}

		// 5. Database: Save the user.
		// We use $1, $2, $3 (placeholders) to prevent "SQL Injection" attacks.
		// We use db.DB.Exec because we are executing an action, not asking for rows back.
		_, err = db.DB.Exec(
			`INSERT INTO users (name, email, password_hash, role)
			VALUES ($1, $2, $3, 'USER')`,
			req.Name, req.Email, string(hashedPassword),
		)

		// 6. Error Handling: Did the insert fail?
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user(check if email exist)"})
			return
		}

		//7. Success

		c.JSON(http.StatusCreated, gin.H{"message": "user registred successfully"})
	})

	r.POST("/auth/login", func(c *gin.Context) {
		c.JSON(501, gin.H{"message": "not implemented"})
	})
}
