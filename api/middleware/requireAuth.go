package middleware

import (
	"errors"
	"net/http"
	"os"

	"github.com/Tokenzrey/FPPBKKGOLANG/db/initializers"
	"github.com/Tokenzrey/FPPBKKGOLANG/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type AuthUser struct {
	ID    uint   `json:"ID"`
	Name  string `json:"Name"`
	Email string `json:"Email"`
}

// GetUserIDFromToken extracts the user ID (sub) from a JWT token.
func GetUserIDFromToken(c *gin.Context) (float64, error) {
	// Extract the Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("missing Authorization header")
	}

	// Ensure the token uses "Bearer" format
	if len(authHeader) <= len("Bearer ") {
		return 0, errors.New("invalid Authorization header format")
	}
	tokenStr := authHeader[len("Bearer "):]

	// Parse the JWT token and validate its signature
	claims := jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil {
		return 0, errors.New("invalid or expired token")
	}

	// Extract the user ID ("sub") from the claims
	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("user ID not found in token payload")
	}

	return userID, nil
}

// RequireAuth is a middleware to check for user authentication and attach user info to context.
func RequireAuth(c *gin.Context) {
	// Extract user ID from the token
	userID, err := GetUserIDFromToken(c)
	if err != nil {
		// Respond with unauthorized if token is missing, invalid, or expired
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Find the user in the database using the extracted user ID
	var user models.User
	if err := initializers.DB.First(&user, userID).Error; err != nil || user.ID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Continue to the next middleware or handler
	c.Next()
}
