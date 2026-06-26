package middleware

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"gateway/audit"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	auditChannelSize = 512
	auditBodyLimit   = 64 * 1024 // 64 KB
)

type auditEntry struct {
	userID     string
	method     string
	url        string
	durationMs int32
	reqBody    string
	respBody   string
	respStatus int32
}

type AuditLogger struct {
	ch     chan auditEntry
	client audit.AuditServiceClient
}

var (
	auditClientOnce sync.Once
	auditClientInst audit.AuditServiceClient
)

func InitAuditClient() audit.AuditServiceClient {
	auditClientOnce.Do(func() {
		addr := os.Getenv("AUDIT_SERVICE_ADDRESS")
		if addr == "" {
			addr = "localhost:50060"
		}
		conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("audit middleware: failed to connect to audit service: %v", err)
			return
		}
		auditClientInst = audit.NewAuditServiceClient(conn)
	})
	return auditClientInst
}

func NewAuditLogger(client audit.AuditServiceClient) *AuditLogger {
	al := &AuditLogger{
		ch:     make(chan auditEntry, auditChannelSize),
		client: client,
	}
	go al.worker()
	return al
}

func (al *AuditLogger) worker() {
	for entry := range al.ch {
		if al.client == nil {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, err := al.client.LogActivity(ctx, &audit.LogActivityRequest{
			UserId:     entry.userID,
			Method:     entry.method,
			Url:        entry.url,
			DurationMs: entry.durationMs,
			ReqBody:    entry.reqBody,
			RespBody:   entry.respBody,
			RespStatus: entry.respStatus,
		})
		cancel()
		if err != nil {
			log.Printf("audit middleware: log activity failed: %v", err)
		}
	}
}

// bodyWriter buffers the response body while still writing to the original writer.
type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (bw *bodyWriter) Write(b []byte) (int, error) {
	if remaining := auditBodyLimit - bw.body.Len(); remaining > 0 {
		if len(b) <= remaining {
			bw.body.Write(b)
		} else {
			bw.body.Write(b[:remaining])
		}
	}
	return bw.ResponseWriter.Write(b)
}

func (al *AuditLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// For write methods: buffer body and restore it for downstream handlers.
		// For read methods: capture query string as req_body.
		var reqBodyStr string
		if isWriteMethod(c.Request.Method) && c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(io.LimitReader(c.Request.Body, auditBodyLimit))
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			reqBodyStr = string(bodyBytes)
		} else if q := c.Request.URL.RawQuery; q != "" {
			reqBodyStr = q
		}

		bw := &bodyWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		durationMs := int32(time.Since(start).Milliseconds())
		userID, _ := c.Get(CtxUserID)
		userIDStr, _ := userID.(string)

		entry := auditEntry{
			userID:     userIDStr,
			method:     c.Request.Method,
			url:        c.Request.URL.Path,
			durationMs: durationMs,
			reqBody:    reqBodyStr,
			respBody:   bw.body.String(),
			respStatus: int32(c.Writer.Status()),
		}

		select {
		case al.ch <- entry:
		default:
			log.Printf("audit middleware: channel full, dropping log for %s %s", entry.method, entry.url)
		}
	}
}
