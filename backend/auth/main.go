package main

import (
	"project-devis-auth/controllers"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	public := r.Group("/api")

	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)
	return r
}

func main() {
	r := setupRouter()

	r.Run(":8080")
}