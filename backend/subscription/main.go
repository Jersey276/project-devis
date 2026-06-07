package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net"

	"project-devis-subscription/actions"
	"project-devis-subscription/services"
	subscriptionGrpc "project-devis-subscription/services/grpc"

	stripe "github.com/stripe/stripe-go/v82"
	"google.golang.org/grpc"
)

//go:embed migrations
var migrationsFS embed.FS

var port = flag.Int("port", 50057, "The server port")

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db, migrationsFS)

	stripe.Key = services.ReadSecret(services.StripeSecretKey, services.StripeSecretKeyFile)
	if stripe.Key == "" {
		log.Println("WARNING: STRIPE_SECRET_KEY not set — payment RPCs will fail")
	}

	webhookSecret := services.ReadSecret(services.StripeWebhookSecret, services.StripeWebhookSecretFile)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	subscriptionGrpc.RegisterSubscriptionServiceServer(grpcServer, actions.NewServer(db, webhookSecret))

	log.Printf("subscription gRPC server listening on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}
