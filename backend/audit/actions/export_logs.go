package actions

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sync"

	auditGrpc "project-devis-audit/services/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	emailClientOnce sync.Once
	emailClient     auditGrpc.EmailServiceClient
)

func getEmailClient() auditGrpc.EmailServiceClient {
	emailClientOnce.Do(func() {
		addr := os.Getenv("EMAIL_SERVICE_ADDRESS")
		if addr == "" {
			addr = "localhost:50058"
		}
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("audit: failed to connect to email service: %v", err)
			return
		}
		emailClient = auditGrpc.NewEmailServiceClient(conn)
	})
	return emailClient
}

func (s *Server) ExportActivityLogs(ctx context.Context, req *auditGrpc.ExportActivityLogsRequest) (*auditGrpc.ExportActivityLogsResponse, error) {
	if req.RecipientEmail == "" {
		return &auditGrpc.ExportActivityLogsResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	where, args := buildFilters(req.Filters)
	query := `SELECT id, COALESCE(user_id,''), method, url, duration_ms,
	                 COALESCE(req_body,''), resp_body, resp_status,
	                 to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"')
	          FROM activity_logs` + where + ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return &auditGrpc.ExportActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
	}
	defer rows.Close()

	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.Write([]string{"id", "user_id", "method", "url", "duration_ms", "req_body", "resp_body", "resp_status", "created_at"})

	for rows.Next() {
		var (
			id, durationMs, respStatus int64
			userID, method, url        string
			reqBody, respBody          string
			createdAt                  string
		)
		if err := rows.Scan(&id, &userID, &method, &url, &durationMs, &reqBody, &respBody, &respStatus, &createdAt); err != nil {
			return &auditGrpc.ExportActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
		}
		_ = w.Write([]string{
			fmt.Sprintf("%d", id), userID, method, url,
			fmt.Sprintf("%d", durationMs), reqBody, respBody,
			fmt.Sprintf("%d", respStatus), createdAt,
		})
	}
	w.Flush()

	client := getEmailClient()
	if client == nil {
		log.Printf("audit export: email service unavailable, CSV not sent to %s", req.RecipientEmail)
		return &auditGrpc.ExportActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
	}

	_, sendErr := client.SendGenericEmail(ctx, &auditGrpc.SendGenericEmailRequest{
		ToEmail:         req.RecipientEmail,
		ToName:          req.RecipientName,
		Subject:         "Export du journal d'activité",
		TextBody:        "Veuillez trouver ci-joint l'export du journal d'activité au format CSV.",
		AttachmentName:  "journal-activite.csv",
		AttachmentBytes: buf.Bytes(),
		AttachmentType:  "text/csv",
	})
	if sendErr != nil {
		log.Printf("audit export: send email failed: %v", sendErr)
		return &auditGrpc.ExportActivityLogsResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &auditGrpc.ExportActivityLogsResponse{Success: true, Code: CodeSuccess}, nil
}
