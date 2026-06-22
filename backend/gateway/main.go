package main

import (
	"gateway/authz"
	"gateway/controllers"
	"gateway/middleware"
	"gateway/services"

	"github.com/gin-gonic/gin"
)

var authorizer = authz.NewFromEnv()

type Route struct {
	TargetURL string
}

func main() {
	auditClient := middleware.InitAuditClient()
	auditLogger := middleware.NewAuditLogger(auditClient)
	r := setupRouter(auditLogger)
	r.Run(":8080")
}

func setupRouter(auditLogger *middleware.AuditLogger) *gin.Engine {
	r := gin.Default()
	r.Use(auditLogger.Middleware())

	emailNotifier := services.NewEmailNotifier()

	// Webhooks — no auth, raw body, registered before auth groups
	webhooks := r.Group("/api/webhooks")
	controllers.WebhookRoutes(webhooks)
	controllers.ResendWebhookRoutes(webhooks)

	api := r.Group("/api")
	controllers.AuthRoutes(api.Group("/auth"))

	users := api.Group("/users")
	users.Use(middleware.AuthRequired())
	controllers.UserRoutes(users)

	quotes := api.Group("/quotes")
	quotes.Use(middleware.AuthRequired())
	controllers.QuotesRoutes(quotes, emailNotifier)

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
	controllers.SchedulesRoutes(schedules, emailNotifier)

	fees := api.Group("/fees")
	fees.Use(middleware.AuthRequired())
	fees.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionFees))
	controllers.FeesRoutes(fees)
	
	invoices := api.Group("/invoices")
	invoices.Use(middleware.AuthRequired())
	invoices.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	controllers.InvoicesRoutes(invoices)

	creditNotes := api.Group("/credit-notes")
	creditNotes.Use(middleware.AuthRequired())
	creditNotes.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	controllers.CreditNotesRoutes(creditNotes)

	plans := api.Group("/plans")
	plans.Use(middleware.AuthRequired())
	controllers.PlansRoutes(plans)

	subscriptions := api.Group("/subscriptions")
	subscriptions.Use(middleware.AuthRequired())
	controllers.SubscriptionsRoutes(subscriptions)

	emailLogs := api.Group("/email-logs")
	emailLogs.Use(middleware.AuthRequired())
	controllers.EmailLogsRoutes(emailLogs, authorizer)

	logs := api.Group("/logs")
	logs.Use(middleware.AuthRequired())
	logs.Use(middleware.RequireSuperAdmin())
	controllers.AuditRoutes(logs)

	return r
}
