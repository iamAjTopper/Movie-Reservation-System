package routes

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"movie-reservation/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
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
		var req loginRequest
		// 1. Parsing: Read the JSON sent by the user
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Prepare variables to hold the data we pull from the database.
		var userID int
		var passwordHash string
		var role string

		// 2. Identification: Find the user by Email.
		// We use QueryRow because we expect exactly ONE user (emails are unique).
		// .Scan() copies the columns from the DB row into our Go variables.
		err := db.DB.QueryRow(
			`SELECT id, password_hash, role FROM users WHERE email = $1`,
			req.Email,
		).Scan(&userID, &passwordHash, &role)

		// 3. Database Check: Did we find them?
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
			return
		}

		// 4. Verification: Check the Password.
		// We CANNOT just compare strings (req.Password == passwordHash) because of the salt.
		// We must use bcrypt's special function to mathematically verify the match.
		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// 5. Authorization: Create the "Key Card" (JWT).
		// Claims are the data written onto the badge.
		claims := jwt.MapClaims{
			"user_id": userID,
			"role":    role,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		}

		// Create the token object using the HS256 signing method.
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// 6. Signing: Seal the token with your secret key.
		// This ensures no one can tamper with the token.
		// Make sure "JWT_SECRET" is set in your environment variables!
		signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token generated"})
			return
		}

		// 7. Success: Hand the token to the user
		c.JSON(http.StatusOK, gin.H{
			"token": signedToken,
		})
	})
}
