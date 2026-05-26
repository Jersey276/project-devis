package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"project-devis-export/actions"
	"project-devis-export/quote"
	"project-devis-export/services/gotenberg"
	exportGrpc "project-devis-export/services/grpc"
	"project-devis-export/users"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var port = flag.Int("port", 50054, "The server port")

// 8 MiB — generous headroom for realistic quote PDFs (typical: 50 KiB–1 MiB).
// If we start embedding heavy media we'll switch to server-streaming gRPC
// rather than raise this further. Mirrored on the gateway client.
const maxExportMessageBytes = 8 * 1024 * 1024

func main() {
	flag.Parse()

	quoteAddr := envOrDefault("QUOTE_SERVICE_ADDRESS", "localhost:50053")
	usersAddr := envOrDefault("USER_SERVICE_ADDRESS", "localhost:50052")
	gotenbergAddr := envOrDefault("GOTENBERG_ADDRESS", "http://localhost:3000")

	qConn, err := grpc.NewClient(quoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial quote service: %v", err)
	}
	uConn, err := grpc.NewClient(usersAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("dial users service: %v", err)
	}

	qClient := quote.NewQuoteServiceClient(qConn)
	uClient := users.NewUserServiceClient(uConn)
	gtClient := gotenberg.New(gotenbergAddr)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxExportMessageBytes),
		grpc.MaxSendMsgSize(maxExportMessageBytes),
	)
	exportGrpc.RegisterExportServiceServer(grpcServer, actions.NewServer(qClient, uClient, gtClient))
	log.Printf("export configured with gotenberg=%s", gotenbergAddr)

	log.Printf("export gRPC server listening on %s", lis.Addr().String())
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
