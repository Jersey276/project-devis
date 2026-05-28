package controllers

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	auth "gateway/auth"
	"gateway/authcookie"

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
	CodeInvalidResetToken   int32 = 1005
	CodeExpiredResetToken   int32 = 1006
	CodeWeakPassword        int32 = 1007
	CodeUserServiceError    int32 = 2001
	CodeInternalError       int32 = 2002
	CodeNotImplemented      int32 = 2003
)

const (
	cookieAccessMaxAge          = 15 * 60          // must not outlive the JWT (services/jwt.go)
	cookieRefreshMaxAge         = 7 * 24 * 60 * 60 // default refresh-token lifetime
	cookieRefreshRememberMaxAge = 60 * 24 * 60 * 60
)

var cookieSecure = os.Getenv("ENV") == "production"

var (
	resetPasswordIPLimiter    = newSlidingWindowLimiter()
	resetPasswordEmailLimiter = newSlidingWindowLimiter()
	confirmResetIPLimiter     = newSlidingWindowLimiter()
)

func setAuthCookies(c *gin.Context, accessToken, refreshToken string, rememberMe bool) {
	refreshMaxAge := cookieRefreshMaxAge
	if rememberMe {
		refreshMaxAge = cookieRefreshRememberMaxAge
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authcookie.AccessName, accessToken, cookieAccessMaxAge, "/", "", cookieSecure, true)
	c.SetCookie(authcookie.RefreshName, refreshToken, refreshMaxAge, "/", "", cookieSecure, true)
}

func clearAuthCookies(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authcookie.AccessName, "", -1, "/", "", cookieSecure, true)
	c.SetCookie(authcookie.RefreshName, "", -1, "/", "", cookieSecure, true)
}

var authErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		CodeUserAlreadyExists:   {http.StatusConflict, "Un compte avec cette adresse email existe déjà."},
		CodeUserNotFound:        {http.StatusNotFound, "Aucun compte trouvé avec cette adresse email."},
		CodeInvalidCredentials:  {http.StatusUnauthorized, "Adresse email ou mot de passe incorrect."},
		CodeInvalidRefreshToken: {http.StatusUnauthorized, "Session expirée, veuillez vous reconnecter."},
		CodeInvalidResetToken:   {http.StatusBadRequest, "Le lien de réinitialisation est invalide ou déjà utilisé."},
		CodeExpiredResetToken:   {http.StatusGone, "Le lien de réinitialisation a expiré."},
		CodeWeakPassword:        {http.StatusUnprocessableEntity, "Le mot de passe ne respecte pas la politique de sécurité."},
		CodeUserServiceError:    {http.StatusBadGateway, "Erreur lors de la création du compte, veuillez réessayer."},
		CodeInternalError:       {http.StatusInternalServerError, "Une erreur interne est survenue."},
		CodeNotImplemented:      {http.StatusNotImplemented, "Cette fonctionnalité n'est pas encore disponible."},
	},
	unavailableMessage: "Service d'authentification indisponible.",
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
	password.POST("/confirm-reset", func(c *gin.Context) { ConfirmResetPassword(c, client) })
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
		authErrors.unavailable(c)
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
		authErrors.reply(c, resp.Code)
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
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.GetCode())
		return
	}

	setAuthCookies(c, resp.GetToken(), resp.GetRefreshToken(), input.RememberMe)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"token":         resp.GetToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

type refreshInput struct {
	RefreshToken string `json:"refresh_token"`
}

func RefreshToken(c *gin.Context, client auth.AuthServiceClient) {
	var input refreshInput
	_ = c.ShouldBindJSON(&input)
	if input.RefreshToken == "" {
		if cookie, err := c.Cookie(authcookie.RefreshName); err == nil {
			input.RefreshToken = cookie
		}
	}
	if input.RefreshToken == "" {
		clearAuthCookies(c)
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Refresh token manquant.", "code": CodeInvalidRefreshToken})
		return
	}

	resp, err := client.RefreshToken(c.Request.Context(), &auth.RefreshTokenRequest{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		clearAuthCookies(c)
		authErrors.reply(c, resp.GetCode())
		return
	}

	setAuthCookies(c, resp.GetToken(), resp.GetRefreshToken(), resp.GetRememberMe())

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"token":         resp.GetToken(),
		"refresh_token": resp.GetRefreshToken(),
	})
}

type logoutInput struct {
	RefreshToken string `json:"refresh_token"`
}

func Logout(c *gin.Context, client auth.AuthServiceClient) {
	var input logoutInput
	_ = c.ShouldBindJSON(&input)
	if input.RefreshToken == "" {
		if cookie, err := c.Cookie(authcookie.RefreshName); err == nil {
			input.RefreshToken = cookie
		}
	}
	// No token to revoke: clear browser state and treat as already logged out.
	if input.RefreshToken == "" {
		clearAuthCookies(c)
		c.JSON(http.StatusOK, gin.H{"success": true, "message": "Déconnexion réussie."})
		return
	}

	resp, err := client.Logout(c.Request.Context(), &auth.LogoutRequest{
		RefreshToken: input.RefreshToken,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	clearAuthCookies(c)
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

	now := time.Now()
	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))
	if !resetPasswordIPLimiter.Allow("ip:"+c.ClientIP(), 5, time.Minute, now) ||
		!resetPasswordEmailLimiter.Allow("email:"+normalizedEmail, 3, 15*time.Minute, now) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "Trop de demandes. Veuillez réessayer plus tard.",
			"code":    "RATE_LIMITED",
		})
		return
	}

	resp, err := client.ResetPassword(c.Request.Context(), &auth.ResetPasswordRequest{
		Email: normalizedEmail,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
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
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mot de passe mis à jour."})
}

func ConfirmResetPassword(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	if !confirmResetIPLimiter.Allow("confirm_ip:"+c.ClientIP(), 10, time.Minute, time.Now()) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "Trop de demandes. Veuillez réessayer plus tard.",
			"code":    "RATE_LIMITED",
		})
		return
	}

	resp, err := client.ConfirmResetPassword(c.Request.Context(), &auth.ConfirmResetPasswordRequest{
		Token:       input.Token,
		NewPassword: input.NewPassword,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Mot de passe réinitialisé."})
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
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Email vérifié."})
}
