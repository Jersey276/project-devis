package main

import (
	"gateway/controllers"

	"github.com/gin-gonic/gin"
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

	api := r.Group("/api")
	controllers.AuthRoutes(api.Group("/auth"))
	// controllers.UserRoutes(api.Group("/users"))
	// controllers.ProjectRoutes(api.Group("/project"))
	// controllers.PaymentRoutes(api.Group("/payments"))

	return r
}