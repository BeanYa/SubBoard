package middleware

import (
	"net/http"
	"strings"

	"submanager/config"
	"submanager/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

// JWTAuth validates JWT token from Authorization header.
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := c.MustGet("config").(*config.Config)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return []byte(cfg.AppSecret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			c.Abort()
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id in token"})
			c.Abort()
			return
		}

		db := c.MustGet("db").(*gorm.DB)
		var user model.User
		if err := db.First(&user, uint(userID)).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		if !user.Enabled {
			c.JSON(http.StatusForbidden, gin.H{"error": "account disabled"})
			c.Abort()
			return
		}

		c.Set("user", &user)
		c.Next()
	}
}
