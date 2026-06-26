package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"

	"gateway/auth"
	"gateway/authcookie"
	"gateway/authz"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	CtxUserID                    = "user_id"
	CtxEmail                     = "email"
	CtxRole                      = "role"
	CtxAccountStatus             = "account_status"
	CtxSubscriptionTier          = "subscription_tier"
	CtxSessionVersion            = "session_version"
	CtxEmailVerified             = "email_verified"
	codeSessionInvalidated int32 = 1008
)

var authorizer = authz.NewFromEnv()

var (
	authClientOnce sync.Once
	authClient     auth.AuthServiceClient
	authClientErr  error
)

func GetAuthServiceClient() (auth.AuthServiceClient, error) {
	return getAuthClient()
}

func getAuthClient() (auth.AuthServiceClient, error) {
	authClientOnce.Do(func() {
		address := os.Getenv("AUTH_SERVICE_ADDRESS")
		if address == "" {
			address = "localhost:50051"
		}
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			authClientErr = err
			return
		}
		authClient = auth.NewAuthServiceClient(conn)
	})
	return authClient, authClientErr
}

func isWriteMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func AuthRequired() gin.HandlerFunc {
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

		client, err := getAuthClient()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Service d'authentification indisponible.",
				"code":    "AUTH_UNAVAILABLE",
			})
			return
		}

		resp, err := client.IntrospectToken(c.Request.Context(), &auth.IntrospectTokenRequest{Token: tokenStr})
		if err != nil || !resp.GetSuccess() || resp.GetContext() == nil {
			statusCode := "TOKEN_INVALID"
			message := "Token invalide ou expiré."
			if resp != nil && resp.GetCode() == codeSessionInvalidated {
				statusCode = "SESSION_INVALIDATED"
				message = "Session expirée, veuillez vous reconnecter."
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": message,
				"code":    statusCode,
			})
			return
		}

		ctx := resp.GetContext()

		c.Set(CtxUserID, ctx.GetUserId())
		c.Set(CtxEmail, ctx.GetEmail())
		c.Set(CtxRole, ctx.GetRole())
		c.Set(CtxAccountStatus, ctx.GetAccountStatus())
		c.Set(CtxSubscriptionTier, ctx.GetSubscriptionTier())
		c.Set(CtxSessionVersion, ctx.GetSessionVersion())
		c.Set(CtxEmailVerified, ctx.GetEmailVerified())

		action := authz.ActionRead
		if isWriteMethod(c.Request.Method) {
			action = authz.ActionManage
		}

		decision, authzErr := authorizer.Can(c.Request.Context(), authz.Subject{
			Role:             ctx.GetRole(),
			AccountStatus:    ctx.GetAccountStatus(),
			SubscriptionTier: ctx.GetSubscriptionTier(),
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
