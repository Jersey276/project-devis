package main

import (
	"gateway/authz"
	"gateway/controllers"
	"gateway/middleware"

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

	// Stripe webhook — no auth, raw body, registered before auth groups
	webhooks := r.Group("/api/webhooks")
	controllers.WebhookRoutes(webhooks)

	api := r.Group("/api")
	controllers.AuthRoutes(api.Group("/auth"))

	users := api.Group("/users")
	users.Use(middleware.AuthRequired())
	controllers.UserRoutes(users)

	quotes := api.Group("/quotes")
	quotes.Use(middleware.AuthRequired())
	controllers.QuotesRoutes(quotes)

	exportGrp := api.Group("/export")
	exportGrp.Use(middleware.AuthRequired())
	controllers.ExportRoutes(exportGrp)

	templates := api.Group("/templates")
	templates.Use(middleware.AuthRequired())
	templates.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionTemplates))
	controllers.TemplateRoutes(templates)

	schedules := api.Group("/schedules")
	schedules.Use(middleware.AuthRequired())
	schedules.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionSchedules))
	controllers.SchedulesRoutes(schedules)

	plans := api.Group("/plans")
	plans.Use(middleware.AuthRequired())
	controllers.PlansRoutes(plans)

	subscriptions := api.Group("/subscriptions")
	subscriptions.Use(middleware.AuthRequired())
	controllers.SubscriptionsRoutes(subscriptions)

	// controllers.ProjectRoutes(api.Group("/project"))
	// controllers.PaymentRoutes(api.Group("/payments"))

	return r
}
