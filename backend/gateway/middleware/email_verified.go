package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RequireEmailVerified() gin.HandlerFunc {
	return func(c *gin.Context) {
		verified, _ := c.Get(CtxEmailVerified)
		if ok, _ := verified.(bool); !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Votre adresse email n'est pas vérifiée.",
				"code":    "EMAIL_NOT_VERIFIED",
			})
			return
		}
		c.Next()
	}
}
