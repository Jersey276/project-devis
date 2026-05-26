package main

import (
	"embed"
	"flag"
	"fmt"
	"log"

	"project-devis-template/actions"
	"project-devis-template/services"

	"github.com/gin-gonic/gin"
)

//go:embed migrations
var migrationsFS embed.FS

var port = flag.Int("port", 8085, "The server port")

func main() {
	flag.Parse()

	db := services.ConnectDB()
	services.RunMigrations(db, migrationsFS)

	r := gin.Default()
	srv := actions.NewServer(db)
	srv.SetupRoutes(r)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("template HTTP server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("serve failed: %v", err)
	}
}
