package controllers

import (
	"log"
	"net/http"
	"os"

	auth "gateway/auth"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Auth service error codes
const (
	CodeSuccess             int32 = 0
	CodeUserAlreadyExists   int32 = 1001
	CodeUserNotFound        int32 = 1002
	CodeInvalidCredentials  int32 = 1003
	CodeInvalidRefreshToken int32 = 1004
	CodeUserServiceError    int32 = 2001
	CodeInternalError       int32 = 2002
	CodeNotImplemented      int32 = 2003
)

// Maps auth service error codes to HTTP status codes and user-facing messages.
var authErrorMap = map[int32]struct {
	Status  int
	Message string
}{
	CodeUserAlreadyExists:   {http.StatusConflict, "Un compte avec cette adresse email existe déjà."},
	CodeUserNotFound:        {http.StatusNotFound, "Aucun compte trouvé avec cette adresse email."},
	CodeInvalidCredentials:  {http.StatusUnauthorized, "Adresse email ou mot de passe incorrect."},
	CodeInvalidRefreshToken: {http.StatusUnauthorized, "Session expirée, veuillez vous reconnecter."},
	CodeUserServiceError:    {http.StatusBadGateway, "Erreur lors de la création du compte, veuillez réessayer."},
	CodeInternalError:       {http.StatusInternalServerError, "Une erreur interne est survenue."},
	CodeNotImplemented:      {http.StatusNotImplemented, "Cette fonctionnalité n'est pas encore disponible."},
}

func authError(c *gin.Context, code int32) {
	if mapped, ok := authErrorMap[code]; ok {
		c.JSON(mapped.Status, gin.H{"success": false, "message": mapped.Message, "code": code})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Une erreur inconnue est survenue.", "code": code})
	}
}

func AuthRoutes(r *gin.RouterGroup) *gin.RouterGroup {
	address := os.Getenv("AUTH_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50051"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth gRPC server: %v", err)
	}
	client := auth.NewAuthServiceClient(conn)

	r.POST("/register", func(c *gin.Context) { Register(c, client) })
	r.POST("/login", func(c *gin.Context) { Login(c, client) })
	r.POST("/refresh", func(c *gin.Context) { RefreshToken(c, client) })
	r.POST("/logout", func(c *gin.Context) { Logout(c, client) })

	password := r.Group("/password")
	password.POST("/reset", func(c *gin.Context) { ResetPassword(c, client) })
	password.POST("/update", func(c *gin.Context) { UpdatePassword(c, client) })

	email := r.Group("/email")
	email.POST("/verify", func(c *gin.Context) { VerifyEmail(c, client) })

	return r
}

type registerInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(c *gin.Context, client auth.AuthServiceClient) {
	var input registerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "SERVICE_UNAVAILABLE"})
		return
	}

	resp, err := client.Register(c.Request.Context(), &auth.RegisterRequest{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		if len(resp.FieldErrors) > 0 {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"success":      false,
				"code":         resp.Code,
				"field_errors": resp.FieldErrors,
			})
			return
		}
		authError(c, resp.Code)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Inscription réussie."})
}

type loginInput struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

func Login(c *gin.Context, client auth.AuthServiceClient) {
	var input loginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.Login(c.Request.Context(), &auth.LoginRequest{
		Email:      input.Email,
		Password:   input.Password,
		RememberMe: input.RememberMe,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.GetCode())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"token":         resp.GetToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshToken(c *gin.Context, client auth.AuthServiceClient) {
	var input refreshInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.RefreshToken(c.Request.Context(), &auth.RefreshTokenRequest{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.GetCode())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"token":         resp.GetToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

type logoutInput struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Logout(c *gin.Context, client auth.AuthServiceClient) {
	var input logoutInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.Logout(c.Request.Context(), &auth.LogoutRequest{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Déconnexion réussie."})
}

func ResetPassword(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.ResetPassword(c.Request.Context(), &auth.ResetPasswordRequest{
		Email: input.Email,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email de réinitialisation envoyé."})
}

func UpdatePassword(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		Email       string `json:"email" binding:"required,email"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.UpdatePassword(c.Request.Context(), &auth.UpdatePasswordRequest{
		Email:       input.Email,
		OldPassword: input.OldPassword,
		NewPassword: input.NewPassword,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mot de passe mis à jour."})
}

func VerifyEmail(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	resp, err := client.VerifyEmail(c.Request.Context(), &auth.VerifyEmailRequest{
		Email: input.Email,
		Token: input.Token,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "message": "Service d'authentification indisponible.", "code": "SERVICE_UNAVAILABLE"})
		return
	}
	if !resp.Success {
		authError(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email vérifié."})
}
