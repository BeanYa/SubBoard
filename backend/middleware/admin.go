package middleware

import (
	"net/http"

	"submanager/model"

	"github.com/gin-gonic/gin"
)

// AdminAuth checks if the authenticated user is an admin.
// Must be used after JWTAuth middleware.
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		val, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}

		user, ok := val.(*model.User)
		if !ok || !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
