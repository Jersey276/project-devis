package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"gateway/authcookie"
	"gateway/authz"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserID           = "user_id"
	CtxEmail            = "email"
	CtxRole             = "role"
	CtxAccountStatus    = "account_status"
	CtxSubscriptionTier = "subscription_tier"
)

type authClaims struct {
	Email            string `json:"email"`
	UserID           string `json:"user_id"`
	Role             string `json:"role"`
	AccountStatus    string `json:"account_status"`
	SubscriptionTier string `json:"subscription_tier"`
	jwt.RegisteredClaims
}

var authorizer = authz.NewFromEnv()

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func AuthRequired() gin.HandlerFunc {
	key := []byte(os.Getenv("APP_KEY"))
	return func(c *gin.Context) {
		var tokenStr string
		if header := c.GetHeader("Authorization"); strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else if cookie, err := c.Cookie(authcookie.AccessName); err == nil {
			tokenStr = cookie
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token d'authentification manquant.",
			})
			return
		}

		token, err := jwt.ParseWithClaims(tokenStr, &authClaims{}, func(t *jwt.Token) (interface{}, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("algorithme de signature inattendu: %v", t.Header["alg"])
			}
			return key, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token invalide ou expiré.",
			})
			return
		}

		claims, ok := token.Claims.(*authClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Token invalide.",
			})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxEmail, claims.Email)
		c.Set(CtxRole, claims.Role)
		c.Set(CtxAccountStatus, claims.AccountStatus)
		c.Set(CtxSubscriptionTier, claims.SubscriptionTier)

		action := authz.ActionRead
		if isWriteMethod(c.Request.Method) {
			action = authz.ActionManage
		}

		decision, authzErr := authorizer.Can(c.Request.Context(), authz.Subject{
			Role:             claims.Role,
			AccountStatus:    claims.AccountStatus,
			SubscriptionTier: claims.SubscriptionTier,
		}, action, authz.ResourceGeneral)
		if authzErr != nil {
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
				"message": "Accès refusé.",
				"code":    decision.Reason,
			})
			return
		}

		c.Next()
	}
}

func RequireSuperAdmin() gin.HandlerFunc {
	return RequireAdminResource(authz.ResourceAdminCountries)
}

func RequireAdminResource(resource authz.Resource) gin.HandlerFunc {
	return func(c *gin.Context) {
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
				"message": "Accès administrateur requis.",
				"code":    decision.Reason,
			})
			return
		}

		c.Next()
	}
}

func subjectFromContext(c *gin.Context) authz.Subject {
	role, _ := c.Get(CtxRole)
	status, _ := c.Get(CtxAccountStatus)
	tier, _ := c.Get(CtxSubscriptionTier)

	roleStr, _ := role.(string)
	statusStr, _ := status.(string)
	tierStr, _ := tier.(string)

	return authz.Subject{
		Role:             roleStr,
		AccountStatus:    statusStr,
		SubscriptionTier: tierStr,
	}
}
