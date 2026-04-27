package main

import (
	"gateway/controllers"
	"gateway/middleware"

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

	users := api.Group("/users")
	users.Use(middleware.AuthRequired())
	controllers.UserRoutes(users)

	quotes := api.Group("/quotes")
	quotes.Use(middleware.AuthRequired())
	controllers.QuotesRoutes(quotes)

	// controllers.ProjectRoutes(api.Group("/project"))
	// controllers.PaymentRoutes(api.Group("/payments"))

	return r
}