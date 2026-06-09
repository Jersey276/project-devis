package services

import (
	"context"
	"log"
	"os"

	emailGrpc "gateway/email"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const maxEmailMessageBytes = 20 * 1024 * 1024 // 20 MB (PDF attachments)

type EmailNotifier interface {
	SendQuoteEmail(ctx context.Context, userID, quoteID, toEmail, toName, quoteName string, pdf []byte) error
	SendScheduleEmail(ctx context.Context, userID, quoteID, toEmail, toName, quoteName, status string) error
}

// ─── gRPC implementation ─────────────────────────────────────────────────────

type grpcEmailNotifier struct {
	client emailGrpc.EmailServiceClient
}

func (n *grpcEmailNotifier) SendQuoteEmail(ctx context.Context, userID, quoteID, toEmail, toName, quoteName string, pdf []byte) error {
	_, err := n.client.SendQuoteEmail(ctx, &emailGrpc.SendQuoteEmailRequest{
		UserId:    userID,
		QuoteId:   quoteID,
		ToEmail:   toEmail,
		ToName:    toName,
		QuoteName: quoteName,
		PdfBytes:  pdf,
	})
	return err
}

func (n *grpcEmailNotifier) SendScheduleEmail(ctx context.Context, userID, quoteID, toEmail, toName, quoteName, status string) error {
	_, err := n.client.SendScheduleEmail(ctx, &emailGrpc.SendScheduleEmailRequest{
		UserId:    userID,
		QuoteId:   quoteID,
		ToEmail:   toEmail,
		ToName:    toName,
		QuoteName: quoteName,
		Status:    status,
	})
	return err
}

// ─── Log fallback ─────────────────────────────────────────────────────────────

type logEmailNotifier struct{}

func (n *logEmailNotifier) SendQuoteEmail(_ context.Context, _, _, toEmail, _, quoteName string, _ []byte) error {
	log.Printf("email fallback: send quote to=%s quote=%q", toEmail, quoteName)
	return nil
}

func (n *logEmailNotifier) SendScheduleEmail(_ context.Context, _, _, toEmail, _, quoteName, status string) error {
	log.Printf("email fallback: schedule notification to=%s quote=%q status=%s", toEmail, quoteName, status)
	return nil
}

// ─── Factory ─────────────────────────────────────────────────────────────────

func NewEmailNotifier() EmailNotifier {
	address := os.Getenv("EMAIL_SERVICE_ADDRESS")
	if address == "" {
		return &logEmailNotifier{}
	}
	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxEmailMessageBytes)),
	)
	if err != nil {
		log.Printf("warning: could not connect to email service at %s: %v — using log fallback", address, err)
		return &logEmailNotifier{}
	}
	return &grpcEmailNotifier{client: emailGrpc.NewEmailServiceClient(conn)}
}
