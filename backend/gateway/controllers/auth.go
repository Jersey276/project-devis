package controllers

import (
	"gateway/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type server struct {
	pb auth.AuthServiceClient
}

func AuthRoutes(conn *grpc.ClientConn, r *gin.RouterGroup) *gin.RouterGroup {
	grpc := auth.NewAuthServiceClient(conn)

	public := r.Group("/auth")

	public.POST("/register", func(c *gin.Context) { Register(c, grpc) })
	public.POST("/login", func(c *gin.Context) { Login(c, grpc) })
	password := public.Group("/password")
	password.POST("/reset", func(c *gin.Context) { ResetPassword(c, grpc) })
	password.POST("/update", func(c *gin.Context) { UpdatePassword(c, grpc) })

	email := public.Group("/email")
	email.POST("/verify", func(c *gin.Context) { VerifyEmail(c, grpc) })
	return r;
}

func Register(c *gin.Context, grpc auth.AuthServiceClient) {
	
	grpc.Register(c.Request.Context(), &auth.RegisterRequest{
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	})
}

func Login(c *gin.Context, grpc auth.AuthServiceClient) {
	grpc.Login(c.Request.Context(), &auth.LoginRequest{
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	})
}

func ResetPassword(c *gin.Context, grpc auth.AuthServiceClient) {
	grpc.ResetPassword(c.Request.Context(), &auth.ResetPasswordRequest{
		Email: c.PostForm("email"),
	})
}

func UpdatePassword(c *gin.Context, grpc auth.AuthServiceClient) {
	grpc.UpdatePassword(c.Request.Context(), &auth.UpdatePasswordRequest{
		Email:       c.PostForm("email"),
		OldPassword: c.PostForm("old_password"),
		NewPassword: c.PostForm("new_password"),
	})
}

func VerifyEmail(c *gin.Context, grpc auth.AuthServiceClient) {
	grpc.VerifyEmail(c.Request.Context(), &auth.VerifyEmailRequest{
		Email: c.PostForm("email"),
		Token: c.PostForm("token"),
	})
}
