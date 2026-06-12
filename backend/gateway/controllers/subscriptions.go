package controllers

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	authpb "gateway/auth"
	"gateway/authz"
	"gateway/middleware"
	sub "gateway/subscription"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	SubCodeNotFound      int32 = 1001
	SubCodeInvalidInput  int32 = 1003
	SubCodeInternalError int32 = 2001
)

var subscriptionErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		SubCodeNotFound:      {http.StatusNotFound, "Plan ou abonnement introuvable."},
		SubCodeInvalidInput:  {http.StatusBadRequest, "Données invalides."},
		SubCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service abonnement indisponible.",
}

var (
	subClientOnce sync.Once
	subClient     sub.SubscriptionServiceClient
	subClientErr  error
)

func getSubscriptionClient() (sub.SubscriptionServiceClient, error) {
	subClientOnce.Do(func() {
		address := os.Getenv("SUBSCRIPTION_SERVICE_ADDRESS")
		if address == "" {
			address = "localhost:50057"
		}
		conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			subClientErr = err
			return
		}
		subClient = sub.NewSubscriptionServiceClient(conn)
	})
	return subClient, subClientErr
}

func SubscriptionsRoutes(r *gin.RouterGroup) {
	client, err := getSubscriptionClient()
	if err != nil {
		panic("failed to connect to subscription gRPC server: " + err.Error())
	}
	r.GET("/me", func(c *gin.Context) { GetMySubscription(c, client) })
	r.POST("/payment-intent", func(c *gin.Context) { CreatePaymentIntent(c, client) })
	r.POST("/cancel", func(c *gin.Context) { CancelSubscription(c, client) })
	r.GET("/admin", middleware.RequireAdminResource(authz.ResourceAdminSubscriptions), func(c *gin.Context) { ListSubscriptionsAdmin(c, client) })
	r.GET("/admin/stats", middleware.RequireAdminResource(authz.ResourceAdminSubscriptions), func(c *gin.Context) { GetAdminStats(c, client) })
	r.POST("/admin/:userId/plan", middleware.RequireAdminResource(authz.ResourceAdminSubscriptions), func(c *gin.Context) { AssignPlan(c, client) })
}

func PlansRoutes(r *gin.RouterGroup) {
	client, err := getSubscriptionClient()
	if err != nil {
		panic("failed to connect to subscription gRPC server: " + err.Error())
	}
	r.GET("", func(c *gin.Context) { ListPlans(c, client) })
	r.PUT("/:planId", middleware.RequireAdminResource(authz.ResourceAdminSubscriptions), func(c *gin.Context) { UpdatePlan(c, client) })
}

func WebhookRoutes(r *gin.RouterGroup) {
	client, err := getSubscriptionClient()
	if err != nil {
		panic("failed to connect to subscription gRPC server: " + err.Error())
	}
	r.POST("/stripe", func(c *gin.Context) { HandleStripeWebhook(c, client) })
}

// ─── Handlers ────────────────────────────────────────────────────────────────

func ListPlans(c *gin.Context, client sub.SubscriptionServiceClient) {
	includeInactive := c.Query("include_inactive") == "true"
	resp, err := client.ListPlans(c.Request.Context(), &sub.ListPlansRequest{IncludeInactive: includeInactive})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Plans))
	for _, p := range resp.Plans {
		out = append(out, gin.H{
			"plan_id":         p.PlanId,
			"name":            p.Name,
			"tier":            p.Tier,
			"price_cents":     p.PriceCents,
			"billing_cycle":   p.BillingCycle,
			"features":        p.Features,
			"active":          p.Active,
			"stripe_price_id": p.StripePriceId,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "plans": out})
}

func UpdatePlan(c *gin.Context, client sub.SubscriptionServiceClient) {
	planID, err := strconv.Atoi(c.Param("planId"))
	if err != nil || planID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID de plan invalide."})
		return
	}

	var input struct {
		Name          string `json:"name"           binding:"required"`
		PriceCents    int32  `json:"price_cents"`
		BillingCycle  string `json:"billing_cycle"  binding:"required"`
		StripePriceId string `json:"stripe_price_id"`
		Features      string `json:"features"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	resp, err := client.UpdatePlan(c.Request.Context(), &sub.UpdatePlanRequest{
		PlanId:        int32(planID),
		Name:          input.Name,
		PriceCents:    input.PriceCents,
		BillingCycle:  input.BillingCycle,
		StripePriceId: input.StripePriceId,
		Features:      input.Features,
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}

	p := resp.Plan
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plan": gin.H{
			"plan_id":         p.PlanId,
			"name":            p.Name,
			"tier":            p.Tier,
			"price_cents":     p.PriceCents,
			"billing_cycle":   p.BillingCycle,
			"features":        p.Features,
			"active":          p.Active,
			"stripe_price_id": p.StripePriceId,
		},
	})
}

func GetMySubscription(c *gin.Context, client sub.SubscriptionServiceClient) {
	resp, err := client.GetUserSubscription(c.Request.Context(), &sub.GetUserSubscriptionRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}
	s := resp.Subscription
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"subscription": gin.H{
			"subscription_id":        s.SubscriptionId,
			"user_id":                s.UserId,
			"plan_id":                s.PlanId,
			"tier":                   s.Tier,
			"status":                 s.Status,
			"current_period_start":   s.CurrentPeriodStart,
			"current_period_end":     s.CurrentPeriodEnd,
			"cancel_at_period_end":   s.CancelAtPeriodEnd,
			"stripe_subscription_id": s.StripeSubscriptionId,
			"updated_at":             s.UpdatedAt,
		},
	})
}

func ListSubscriptionsAdmin(c *gin.Context, client sub.SubscriptionServiceClient) {
	resp, err := client.ListSubscriptions(c.Request.Context(), &sub.ListSubscriptionsRequest{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}
	out := make([]gin.H, 0, len(resp.Subscriptions))
	for _, s := range resp.Subscriptions {
		out = append(out, gin.H{
			"subscription_id":      s.SubscriptionId,
			"user_id":              s.UserId,
			"plan_id":              s.PlanId,
			"tier":                 s.Tier,
			"status":               s.Status,
			"current_period_start": s.CurrentPeriodStart,
			"current_period_end":   s.CurrentPeriodEnd,
			"cancel_at_period_end": s.CancelAtPeriodEnd,
			"updated_at":           s.UpdatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "subscriptions": out, "total": resp.Total})
}

func AssignPlan(c *gin.Context, client sub.SubscriptionServiceClient) {
	var input struct {
		PlanID int32 `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	userID := c.Param("userId")
	resp, err := client.AssignPlan(c.Request.Context(), &sub.AssignPlanRequest{
		UserId: userID,
		PlanId: input.PlanID,
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}

	authClient, authClientErr := middleware.GetAuthServiceClient()
	if authClientErr != nil {
		log.Printf("AssignPlan: failed to get auth client: %v", authClientErr)
	} else {
		if _, authErr := authClient.UpdateSubscriptionTier(c.Request.Context(), &authpb.UpdateSubscriptionTierRequest{
			UserId: userID,
			Tier:   resp.NewTier,
		}); authErr != nil {
			log.Printf("AssignPlan: failed to update subscription tier in auth for user %s: %v", userID, authErr)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

func CreatePaymentIntent(c *gin.Context, client sub.SubscriptionServiceClient) {
	var input struct {
		PlanID int32 `json:"plan_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	email, _ := c.Get(middleware.CtxEmail)
	emailStr, _ := email.(string)

	resp, err := client.CreatePaymentIntent(c.Request.Context(), &sub.CreatePaymentIntentRequest{
		UserId: userIDFromCtx(c),
		PlanId: input.PlanID,
		Email:  emailStr,
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":                true,
		"client_secret":          resp.ClientSecret,
		"stripe_subscription_id": resp.StripeSubscriptionId,
	})
}

func HandleStripeWebhook(c *gin.Context, client sub.SubscriptionServiceClient) {
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Corps de requête invalide."})
		return
	}

	sig := c.GetHeader("Stripe-Signature")
	resp, err := client.HandleStripeWebhook(c.Request.Context(), &sub.HandleStripeWebhookRequest{
		Payload:         body,
		StripeSignature: sig,
	})
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false})
		return
	}
	if !resp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "code": resp.Code})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func CancelSubscription(c *gin.Context, client sub.SubscriptionServiceClient) {
	resp, err := client.CancelSubscription(c.Request.Context(), &sub.CancelSubscriptionRequest{
		UserId: userIDFromCtx(c),
	})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func GetAdminStats(c *gin.Context, client sub.SubscriptionServiceClient) {
	resp, err := client.GetAdminStats(c.Request.Context(), &sub.GetAdminStatsRequest{})
	if err != nil {
		subscriptionErrors.unavailable(c)
		return
	}
	if !resp.Success {
		subscriptionErrors.reply(c, resp.Code)
		return
	}

	planDist := make([]gin.H, 0, len(resp.PlanDistribution))
	for _, e := range resp.PlanDistribution {
		planDist = append(planDist, gin.H{"tier": e.Tier, "count": e.Count})
	}
	monthly := make([]gin.H, 0, len(resp.MonthlyRevenue))
	for _, e := range resp.MonthlyRevenue {
		monthly = append(monthly, gin.H{"month": e.Month, "revenue_cents": e.RevenueCents})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":                    true,
		"total_active_subscriptions": resp.TotalActiveSubscriptions,
		"total_revenue_cents":        resp.TotalRevenueCents,
		"plan_distribution":          planDist,
		"monthly_revenue":            monthly,
	})
}
