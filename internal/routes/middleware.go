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
		// 1. Check if they are even holding a badge.
		// We look for a specific header called "Authorization".
		authHeader := c.GetHeader("Authorization")
		// If the header is empty, stop them immediately.
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing tokken"})
			return
		}

		// 2. Check the format.
		// The standard format is "Bearer <token>".
		// We split the string by space into two parts.

		parts := strings.Split(authHeader, " ")

		// If it doesn't have 2 parts, or the first word isn't "Bearer", it's bad.
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// The actual token is the second part (index 1)
		tokenStr := parts[1]

		// 3. Verify the Badge (The hardest part).
		// jwt.Parse does two things:
		//    a. It reads the token.
		//    b. It asks YOU for the secret key to verify the signature.

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		// If Parse failed (key didn't match) or the token is expired (!token.Valid).
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// 4. Extract the Info (The Claims).
		// We successfully unlocked the token! Now let's read the data inside.

		claims := token.Claims.(jwt.MapClaims)

		userID := int(claims["user_id"].(float64))
		role := claims["role"].(string)

		// 5. Pin the Name Tag.
		// c.Set saves this data into the request context.
		// Now, any function running AFTER this middleware can say c.Get("user_id")
		// to know exactly who is logged in
		c.Set("user_id", userID)
		c.Set("role", role)

		// 6. Open the Gate.
		// c.Next() tells Gin: "This request is safe. Move to the next function."
		c.Next()
	}
}

// onlt admin middleware
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Retrieve the role from the "Backpack" (Context).
		// Remember: The AuthMiddleware ran BEFORE this.
		// It already unpacked the token and put "role" into 'c'.
		role, exists := c.Get("role")

		// 2. The Double Check.
		// First: "Did the previous middleware even run?" (!exists)
		// Second: "Is the role exactly 'ADMIN'?" (role != "ADMIN")
		if !exists || role != "ADMIN" {
			// 3. The Rejection.
			// 403 Forbidden means: "I know who you are, but you aren't allowed here."
			// (Compare this to 401 Unauthorized, which means "I don't know who you are".)

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		// 4. The Approval.
		// If they are an ADMIN, let them pass to the next function.
		c.Next()

	}
}
