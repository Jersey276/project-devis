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
	"project-devis-invoice/pdp/iopole"
	exportGrpc "project-devis-invoice/services/exportgrpc"
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

	// PA (Plateforme Agréée) adapter selection (B6). Default = no-op (no real PA,
	// no network call). PDP_PROVIDER=iopole wires the real Iopole adapter, which
	// deposits the Factur-X PDF/A-3 fetched from the export service.
	pdpClient, pdpDirectory, reporter := buildPDPAdapters()

	server := actions.NewServer(db, qClient, uClient, sClient, pdpClient, pdpDirectory, reporter)

	// B6 status poller: reconciles deposited invoices' lifecycle with the PA.
	// Disabled by default (interval 0) — inert with the no-op adapter; set
	// PDP_POLL_INTERVAL once a real PA is wired. Stops when the process exits.
	if interval := parseDurationOrZero("PDP_POLL_INTERVAL"); interval > 0 {
		go server.StartPDPPoller(context.Background(), interval, pdpPollSweepTimeout)
		log.Printf("invoice pdp poller enabled (interval=%s)", interval)
	}

	// B5/C5 e-reporting poller: reconciles submitted period aggregates' status with
	// the PA. Same gating philosophy as the PDP poller — disabled by default, inert
	// with the no-op reporter; set REPORT_POLL_INTERVAL once a real PA is wired.
	if interval := parseDurationOrZero("REPORT_POLL_INTERVAL"); interval > 0 {
		go server.StartReportPoller(context.Background(), interval, pdpPollSweepTimeout)
		log.Printf("invoice report poller enabled (interval=%s)", interval)
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

// buildPDPAdapters selects the PA client+directory+reporter from PDP_PROVIDER.
// Anything but "iopole" (including unset) keeps the inert no-op adapters, so
// production is unchanged until a provider is explicitly configured. The Iopole
// client deposits the Factur-X PDF/A-3 obtained from the export service. The
// e-reporting adapter (B5/C5) has no real Iopole implementation yet, so it stays
// no-op even under PDP_PROVIDER=iopole (deferred to a later iteration).
func buildPDPAdapters() (pdp.Client, pdp.Directory, pdp.Reporter) {
	switch envOrDefault("PDP_PROVIDER", "noop") {
	case "iopole":
		baseURL := envOrDefault("IOPOLE_BASE_URL", "https://api.ppd.iopole.fr/v1/api")
		tokenURL := envOrDefault("IOPOLE_TOKEN_URL",
			"https://auth.preprod.iopole.fr/realms/iopole/protocol/openid-connect/token")
		clientID := os.Getenv("IOPOLE_CLIENT_ID")
		clientSecret := os.Getenv("IOPOLE_CLIENT_SECRET")
		customerID := os.Getenv("IOPOLE_CUSTOMER_ID")

		exportAddr := envOrDefault("EXPORT_SERVICE_ADDRESS", "localhost:50054")
		eConn, err := grpc.NewClient(exportAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("dial export service (iopole document source): %v", err)
		}
		docs := actions.NewExportDocumentSource(exportGrpc.NewExportServiceClient(eConn))

		log.Printf("invoice PA adapter: iopole (base=%s); e-reporting adapter: noop (not yet implemented)", baseURL)
		return iopole.NewClient(baseURL, tokenURL, clientID, clientSecret, customerID, docs),
			iopole.NewDirectory(baseURL, tokenURL, clientID, clientSecret, customerID),
			pdp.NoopReporter{}
	default:
		return pdp.NoopClient{}, pdp.NoopDirectory{}, pdp.NoopReporter{}
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
