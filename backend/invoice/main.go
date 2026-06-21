package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	_ "time/tzdata"

	"project-devis-invoice/actions"
	"project-devis-invoice/pdp"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"

	"project-devis-invoice/services"
	invoiceGrpc "project-devis-invoice/services/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:embed migrations
var migrationsFS embed.FS

var port = flag.Int("port", 50059, "The server port")

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db, migrationsFS)

	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		log.Fatalf("seal backfill: %v", err)
	}

	quoteAddr := envOrDefault("QUOTE_SERVICE_ADDRESS", "localhost:50053")
	usersAddr := envOrDefault("USER_SERVICE_ADDRESS", "localhost:50052")
	scheduleAddr := envOrDefault("SCHEDULE_SERVICE_ADDRESS", "localhost:50056")

	qConn, err := grpc.NewClient(quoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial quote service: %v", err)
	}
	uConn, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial users service: %v", err)
	}
	sConn, err := grpc.NewClient(scheduleAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial schedule service: %v", err)
	}

	qClient := quoteGrpc.NewQuoteServiceClient(qConn)
	uClient := usersGrpc.NewUserServiceClient(uConn)
	sClient := scheduleGrpc.NewScheduleServiceClient(sConn)

	// No PA contracted yet (B6): default to the no-op adapter. The address is read
	// as a forward seam; a real adapter is wired here once a provider is chosen.
	_ = envOrDefault("PDP_SERVICE_ADDRESS", "")
	var pdpClient pdp.Client = pdp.NoopClient{}

	server := actions.NewServer(db, qClient, uClient, sClient, pdpClient)

	// B6 status poller: reconciles deposited invoices' lifecycle with the PA.
	// Disabled by default (interval 0) — inert with the no-op adapter; set
	// PDP_POLL_INTERVAL once a real PA is wired. Stops when the process exits.
	if interval := parseDurationOrZero("PDP_POLL_INTERVAL"); interval > 0 {
		go server.StartPDPPoller(context.Background(), interval, pdpPollSweepTimeout)
		log.Printf("invoice pdp poller enabled (interval=%s)", interval)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	invoiceGrpc.RegisterInvoiceServiceServer(grpcServer, server)

	log.Printf("invoice gRPC server listening on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// pdpPollSweepTimeout bounds one poller sweep so a stuck platform call cannot
// wedge the worker between ticks.
const pdpPollSweepTimeout = 2 * time.Minute

// parseDurationOrZero reads a Go duration from env (e.g. "30s", "5m"); returns 0
// when unset or unparseable, which leaves the poller disabled.
func parseDurationOrZero(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Printf("invalid %s=%q, poller disabled: %v", key, v, err)
		return 0
	}
	return d
}
