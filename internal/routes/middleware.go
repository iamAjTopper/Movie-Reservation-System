package routes

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Check if they are even holding a badge
		// We look for a specific header called "Authorization"
		authHeader := c.GetHeader("Authorization")
		// If the header is empty, stop them immediately
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing tokken"})
			return
		}

		// 2. Check the format
		// The standard format is "Bearer <token>"
		// We split the string by space into two parts

		parts := strings.Split(authHeader, " ")

		// If it doesn't have 2 parts, or the first word isn't "Bearer", it's bad
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// The actual token is the second part (index 1)
		tokenStr := parts[1]

		// 3 Verifying the Badge
		// jwt.Parse does two things
		//    a. It reads the token
		//    b. It asks YOU for the secret key to verify the signature

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// If Parse failed (key didn't match) or the token is expired (!token.Valid)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// 4 Extract the Info (The Claims)
		// We successfully unlocked the token! Now let's read the data inside

		claims := token.Claims.(jwt.MapClaims)

		userID := int(claims["user_id"].(float64))
		role := claims["role"].(string)

		// 5 Pin the Name Tag
		c.Set("user_id", userID)
		c.Set("role", role)

		//This request is safe. Move to the next function."
		c.Next()
	}
}

// only admin middleware
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1 Retrieving the role from the Context

		role, exists := c.Get("role")

		// 2 The Double Check
		// First: "Did the previous middleware even run?" (!exists)
		// Second: "Is the role exactly 'ADMIN'?" (role != "ADMIN")
		if !exists || role != "ADMIN" {
			// 3 The Rejection

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		// 4 The Approval
		// If they are an ADMIN, let them pass to the next function.
		c.Next()

	}
}
