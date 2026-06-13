package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	// Embed the IANA tz database so time.LoadLocation("Europe/Paris") works on
	// the scratch runtime image (no system tzdata). The invoice numbering year
	// is fixed in Paris time, so this must be correct around midnight on Dec 31.
	_ "time/tzdata"

	"project-devis-invoice/actions"
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

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	invoiceGrpc.RegisterInvoiceServiceServer(grpcServer, actions.NewServer(db, qClient, uClient, sClient))

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
