package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User-ID no proporcionado"})
			return
		}

		// Puedes agregar lógica adicional, como validar el tweetID
		c.Set("userID", userID)
		c.Next()
	}
}
