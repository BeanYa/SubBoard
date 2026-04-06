package middleware

import (
	"net/http"

	"submanager/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TokenAuth validates subscription token from URL parameter.
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Param("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing subscription token"})
			c.Abort()
			return
		}

		db := c.MustGet("db").(*gorm.DB)
		var user model.User
		if err := db.Where("sub_token = ?", token).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid subscription token"})
			c.Abort()
			return
		}

		if !user.Enabled {
			c.String(http.StatusForbidden, "# Account disabled\n")
			c.Abort()
			return
		}

		c.Set("sub_user", &user)
		c.Next()
	}
}
