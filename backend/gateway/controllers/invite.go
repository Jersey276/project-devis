package controllers

import (
	"log"
	"net/http"
	"os"
	"time"

	authpb "gateway/auth"
	"gateway/middleware"
	users "gateway/users"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var acceptInvitationLimiter = newSlidingWindowLimiter()

// InviteRoutes registers the client invitation endpoints.
// The router group is expected to already be under /api/auth/invite.
func InviteRoutes(r *gin.RouterGroup, authClient authpb.AuthServiceClient, usersClient users.UserServiceClient) {
	// Send invitation — provider must be authenticated.
	send := r.Group("/client")
	send.Use(middleware.AuthRequired())
	send.POST("", func(c *gin.Context) { SendClientInvitation(c, authClient, usersClient) })

	// Accept invitation — public (no existing session required).
	r.POST("/accept", func(c *gin.Context) { AcceptClientInvitationNew(c, authClient) })

	// Accept invitation — existing authenticated user links their account.
	linked := r.Group("/accept-linked")
	linked.Use(middleware.AuthRequired())
	linked.POST("", func(c *gin.Context) { AcceptClientInvitationLinked(c, authClient) })
}

type sendInvitationInput struct {
	ClientID string `json:"client_id" binding:"required"`
}

func SendClientInvitation(c *gin.Context, authClient authpb.AuthServiceClient, usersClient users.UserServiceClient) {
	var input sendInvitationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	providerID := userIDFromCtx(c)

	// Fetch the client to get email and name — do not trust the frontend to send these.
	clientResp, err := usersClient.GetClient(c.Request.Context(), &users.GetClientRequest{
		ClientId: input.ClientID,
		UserId:   providerID,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !clientResp.Success {
		usersErrors.reply(c, clientResp.Code)
		return
	}

	cl := clientResp.Client
	if cl.Email == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Ce client n'a pas d'adresse email."})
		return
	}

	clientName := cl.FirstName + " " + cl.LastName

	resp, err := authClient.SendClientInvitation(c.Request.Context(), &authpb.SendClientInvitationRequest{
		ProviderUserId: providerID,
		ClientId:       input.ClientID,
		ClientEmail:    cl.Email,
		ClientName:     clientName,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

type acceptInvitationNewInput struct {
	Token    string `json:"token" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func AcceptClientInvitationNew(c *gin.Context, authClient authpb.AuthServiceClient) {
	ip := c.ClientIP()
	if !acceptInvitationLimiter.Allow(ip, 10, time.Minute, time.Now()) {
		c.JSON(http.StatusTooManyRequests, gin.H{"success": false, "message": "Trop de tentatives. Veuillez réessayer plus tard."})
		return
	}

	var input acceptInvitationNewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	resp, err := authClient.AcceptClientInvitationNew(c.Request.Context(), &authpb.AcceptClientInvitationNewRequest{
		Token:    input.Token,
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	setAuthCookies(c, resp.Token, resp.RefreshToken, false)
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"is_new_account": resp.IsNewAccount,
		"token":          resp.Token,
		"refresh_token":  resp.RefreshToken,
	})
}

type acceptInvitationLinkedInput struct {
	Token string `json:"token" binding:"required"`
}

func AcceptClientInvitationLinked(c *gin.Context, authClient authpb.AuthServiceClient) {
	var input acceptInvitationLinkedInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	resp, err := authClient.AcceptClientInvitationLinked(c.Request.Context(), &authpb.AcceptClientInvitationLinkedRequest{
		Token:  input.Token,
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		authErrors.unavailable(c)
		return
	}
	if !resp.Success {
		authErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func NewInviteAuthClient() authpb.AuthServiceClient {
	address := os.Getenv("AUTH_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50051"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to auth gRPC server for invite routes: %v", err)
	}
	return authpb.NewAuthServiceClient(conn)
}

func NewInviteUsersClient() users.UserServiceClient {
	address := os.Getenv("USER_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50052"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to users gRPC server for invite routes: %v", err)
	}
	return users.NewUserServiceClient(conn)
}
