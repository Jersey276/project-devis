package controllers

import (
	"net/http"
	"strings"
	"time"

	auth "gateway/auth"
	"gateway/middleware"

	"github.com/gin-gonic/gin"
)

const (
	CodeInvalidEmailChangeToken int32 = 1019
	CodeExpiredEmailChangeToken int32 = 1020
	CodeEmailAlreadyInUse       int32 = 1021
)

var (
	requestEmailChangeLimiter = newSlidingWindowLimiter()
	confirmEmailChangeLimiter = newSlidingWindowLimiter()
)

func RequestEmailChange(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		NewEmail string `json:"new_email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	userIDRaw, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token d'authentification manquant."})
		return
	}
	userID, _ := userIDRaw.(string)

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.NewEmail))
	now := time.Now()
	if !requestEmailChangeLimiter.Allow("email_change_req:"+userID, 3, 15*time.Minute, now) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "Trop de demandes. Veuillez réessayer plus tard.",
			"code":    "RATE_LIMITED",
		})
		return
	}

	resp, err := client.RequestEmailChange(c.Request.Context(), &auth.RequestEmailChangeRequest{
		UserId:   userID,
		NewEmail: normalizedEmail,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		switch resp.Code {
		case CodeEmailAlreadyInUse:
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Cette adresse email est déjà utilisée.", "code": resp.Code})
		default:
			authErrors.reply(c, resp.Code)
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Un email de confirmation a été envoyé à la nouvelle adresse."})
}

func ConfirmEmailChange(c *gin.Context, client auth.AuthServiceClient) {
	var input struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides.", "code": "VALIDATION_ERROR"})
		return
	}

	if !confirmEmailChangeLimiter.Allow("email_change_confirm:"+c.ClientIP(), 10, time.Minute, time.Now()) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"message": "Trop de demandes. Veuillez réessayer plus tard.",
			"code":    "RATE_LIMITED",
		})
		return
	}

	resp, err := client.ConfirmEmailChange(c.Request.Context(), &auth.ConfirmEmailChangeRequest{
		Token: input.Token,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		switch resp.Code {
		case CodeInvalidEmailChangeToken:
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Le lien de confirmation est invalide ou déjà utilisé.", "code": resp.Code})
		case CodeExpiredEmailChangeToken:
			c.JSON(http.StatusGone, gin.H{"success": false, "message": "Le lien de confirmation a expiré.", "code": resp.Code})
		case CodeEmailAlreadyInUse:
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "Cette adresse email est déjà utilisée.", "code": resp.Code})
		default:
			authErrors.reply(c, resp.Code)
		}
		return
	}

	// Session is invalidated server-side (session_version bumped); clear cookies
	// so the browser redirects to login with the new email.
	clearAuthCookies(c)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Adresse email mise à jour. Veuillez vous reconnecter."})
}
