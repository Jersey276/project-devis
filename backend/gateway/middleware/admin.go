package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"gateway/users"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const adminLookupTimeout = 3 * time.Second

var (
	adminClientOnce sync.Once
	adminClient     users.UserServiceClient
	adminClientErr  error
)

func adminUserClient() (users.UserServiceClient, error) {
	adminClientOnce.Do(func() {
		address := os.Getenv("USER_SERVICE_ADDRESS")
		if address == "" {
			address = "localhost:50052"
		}
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			adminClientErr = err
			return
		}
		adminClient = users.NewUserServiceClient(conn)
	})
	return adminClient, adminClientErr
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		client, err := adminUserClient()
		if err != nil {
			log.Printf("admin middleware: failed to init user client: %v", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "Service utilisateurs indisponible."})
			return
		}

		userID, _ := c.Get(CtxUserID)
		userIDStr, _ := userID.(string)
		if userIDStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token invalide."})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), adminLookupTimeout)
		defer cancel()

		resp, err := client.GetUserAccessInfo(ctx, &users.GetUserAccessInfoRequest{UserId: userIDStr})
		if err != nil {
			log.Printf("admin middleware: access info lookup failed for %s: %v", userIDStr, err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "Service utilisateurs indisponible."})
			return
		}
		if !resp.GetSuccess() || resp.GetSuspended() || resp.GetRole() != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "Accès interdit."})
			return
		}

		c.Set(CtxEmail, resp.GetEmail())
		c.Next()
	}
}
