package controllers

import (
	"gateway/auth"

	"github.com/gin-gonic/gin"
)

type server struct {
	pb auth.AuthServiceClient
}

func AuthRoutes() *gin.Engine {
	r := gin.Default()
	public := r.Group("/auth")

	public.POST("/register", Register)
	public.POST("/login", Login)
	password := public.Group("/password")
	password.POST("/reset", ResetPassword)
	password.POST("/update", UpdatePassword)

	email := public.Group("/email")
	email.POST("/verify", VerifyEmail)
	return r
}

func Register(c *gin.Context) {
	grpc := auth.NewGRPCClient()

	grpc.Register(c)
}

func Login(c *gin.Context) {
	grpc := auth.NewGRPCClient()

	grpc.Login(c)
}

func ResetPassword(c *gin.Context) {}

func UpdatePassword(c *gin.Context) {}

func VerifyEmail(c *gin.Context) {}