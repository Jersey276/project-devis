package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"gateway/authz"
	emailGrpc "gateway/email"
	"gateway/middleware"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func EmailLogsRoutes(r *gin.RouterGroup, authorizer authz.Authorizer) {
	address := os.Getenv("EMAIL_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50058"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("warning: could not connect to email service: %v", err)
	}
	var client emailGrpc.EmailServiceClient
	if conn != nil {
		client = emailGrpc.NewEmailServiceClient(conn)
	}

	r.GET("", func(c *gin.Context) { GetEmailLogs(c, client, authorizer) })
}

func GetEmailLogs(c *gin.Context, client emailGrpc.EmailServiceClient, authorizer authz.Authorizer) {
	if client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "message": "Service email indisponible."})
		return
	}

	userID := userIDFromCtx(c)

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	resp, err := client.GetEmailLogs(c.Request.Context(), &emailGrpc.GetEmailLogsRequest{
		UserId: userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Erreur lors de la récupération des logs."})
		return
	}

	// Mask tracking fields for free-tier users.
	tierRaw, _ := c.Get(middleware.CtxSubscriptionTier)
	tier, _ := tierRaw.(string)
	subject := authz.Subject{SubscriptionTier: tier}
	trackingDecision, _ := authorizer.Can(c.Request.Context(), subject, authz.ActionRead, authz.ResourceSubscriptionEmailTracking)

	logs := make([]gin.H, 0, len(resp.Logs))
	for _, l := range resp.Logs {
		entry := gin.H{
			"id":             l.Id,
			"to_email":       l.ToEmail,
			"type":           l.Type,
			"reference_name": l.ReferenceName,
			"status":         l.Status,
			"created_at":     l.CreatedAt,
		}
		if trackingDecision.Allowed {
			entry["opened"] = l.Opened
			entry["clicked"] = l.Clicked
		}
		logs = append(logs, entry)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"total":   resp.Total,
	})
}
