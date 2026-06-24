package middleware

import (
	"net/http"

	"gateway/authz"

	"github.com/gin-gonic/gin"
)

func RequireSubscriptionFeature(resource authz.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Requests made in client-view mode bypass the subscription check: the
		// client is only reading their own data and does not hold a subscription.
		if c.GetHeader("X-Client-Mode") == "customer" {
			c.Next()
			return
		}

		decision, err := authorizer.Can(c.Request.Context(), subjectFromContext(c), authz.ActionManage, resource)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Autorisation indisponible.",
				"code":    "AUTHZ_UNAVAILABLE",
			})
			return
		}

		if !decision.Allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "Cette fonctionnalité est réservée aux abonnés Pro et Enterprise.",
				"code":    decision.Reason,
			})
			return
		}

		c.Next()
	}
}
