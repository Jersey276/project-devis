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
    api.Group("/auth", controllers.AuthRoutes())

    return r
}