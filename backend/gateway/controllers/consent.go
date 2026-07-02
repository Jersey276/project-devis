package controllers

import (
	"net/http"

	"gateway/middleware"
	users "gateway/users"

	"github.com/gin-gonic/gin"
)

func ConsentRoutes(r *gin.RouterGroup, client users.UserServiceClient) {
	grp := r.Group("/consent")
	grp.Use(middleware.AuthRequired())
	grp.POST("", func(c *gin.Context) { acceptConsent(c, client) })
	grp.GET("/status", func(c *gin.Context) { getConsentStatus(c, client) })
}

func acceptConsent(c *gin.Context, client users.UserServiceClient) {
	var input struct {
		Type    string `json:"type" binding:"required"`
		Version string `json:"version" binding:"required"`
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

	resp, err := client.AcceptConsent(c.Request.Context(), &users.AcceptConsentRequest{
		UserId:  userID,
		Type:    input.Type,
		Version: input.Version,
		Ip:      c.ClientIP(),
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true})
}

func getConsentStatus(c *gin.Context, client users.UserServiceClient) {
	userIDRaw, exists := c.Get(middleware.CtxUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token d'authentification manquant."})
		return
	}
	userID, _ := userIDRaw.(string)

	resp, err := client.GetConsentStatus(c.Request.Context(), &users.GetConsentStatusRequest{
		UserId: userID,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}

	type entry struct {
		Type       string `json:"type"`
		Version    string `json:"version"`
		AcceptedAt string `json:"accepted_at"`
	}
	entries := make([]entry, len(resp.Consents))
	for i, e := range resp.Consents {
		entries[i] = entry{Type: e.Type, Version: e.Version, AcceptedAt: e.AcceptedAt}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "consents": entries})
}
