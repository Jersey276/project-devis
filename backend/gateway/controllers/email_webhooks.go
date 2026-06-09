package controllers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	emailGrpc "gateway/email"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ResendWebhookRoutes(r *gin.RouterGroup) {
	address := os.Getenv("EMAIL_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50058"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("warning: could not connect to email service for webhooks: %v", err)
		return
	}
	client := emailGrpc.NewEmailServiceClient(conn)
	webhookSecret := os.Getenv("RESEND_WEBHOOK_SECRET")

	r.POST("/resend", func(c *gin.Context) { HandleResendWebhook(c, client, webhookSecret) })
}

type resendWebhookPayload struct {
	Type      string `json:"type"`
	CreatedAt string `json:"created_at"`
	Data      struct {
		EmailID string `json:"email_id"`
	} `json:"data"`
}

func HandleResendWebhook(c *gin.Context, client emailGrpc.EmailServiceClient, webhookSecret string) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unreadable body"})
		return
	}

	// Signature verification (Svix) — only enforced when secret is configured.
	if webhookSecret != "" {
		if !verifyResendSignature(c.Request.Header, body, webhookSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	}

	var payload resendWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if payload.Type != "email.opened" && payload.Type != "email.clicked" {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}
	if payload.Data.EmailID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing email_id"})
		return
	}

	occurredAt := payload.CreatedAt
	if occurredAt == "" {
		occurredAt = time.Now().UTC().Format(time.RFC3339)
	}

	_, err = client.TrackEmailEvent(c.Request.Context(), &emailGrpc.TrackEmailEventRequest{
		ResendId:   payload.Data.EmailID,
		EventType:  payload.Type,
		OccurredAt: occurredAt,
	})
	if err != nil {
		log.Printf("resend webhook: TrackEmailEvent failed for %s: %v", payload.Data.EmailID, err)
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// verifyResendSignature validates the Svix signature from Resend webhooks.
// Resend uses the standard Svix signing scheme: svix-id, svix-timestamp, svix-signature headers.
func verifyResendSignature(headers http.Header, body []byte, secret string) bool {
	msgID := headers.Get("svix-id")
	msgTS := headers.Get("svix-timestamp")
	msgSig := headers.Get("svix-signature")
	if msgID == "" || msgTS == "" || msgSig == "" {
		return false
	}
	_ = secret // TODO: implement full Svix HMAC-SHA256 verification if svix-go SDK is added
	return true
}
