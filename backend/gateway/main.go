package main

import (
	"gateway/controllers"
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)


type Route struct {
    TargetURL string
}

func main() {
    	r := setupRouter()

    r.Run(":8080")
}

func setupRouter() *gin.Engine {
    r := gin.Default()

    conn, err := grpc.NewClient("localhost:50051")
    if err != nil {
        log.Fatalf("Failed to connect to gRPC server: %v", err)
    }
    defer conn.Close()

    api := r.Group("/api")
    authRoutes := api.Group("/auth")
    controllers.AuthRoutes(conn, authRoutes)
    // usersRoutes := api.Group("/users")
    // controllers.UserRoutes(conn, usersRoutes)
    // projectRoutes := api.Group("/project")
    // controllers.ProjectRoutes(conn, projectRoutes)
    // paymentRoutes := api.Group("/payments")
    // controllers.PaymentRoutes(conn, paymentRoutes)


    return r
}