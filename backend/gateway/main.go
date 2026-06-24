package main

import (
	"gateway/audit"
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
	r := setupRouter(auditLogger, auditClient)
	r.Run(":8080")
}

func setupRouter(auditLogger *middleware.AuditLogger, auditClient audit.AuditServiceClient) *gin.Engine {
	r := gin.Default()

	emailNotifier := services.NewEmailNotifier()

	// Webhooks — no auth, raw body, registered before auth groups
	webhooks := r.Group("/api/webhooks")
	controllers.WebhookRoutes(webhooks)
	controllers.ResendWebhookRoutes(webhooks)

	audited := r.Group("/api")
	audited.Use(auditLogger.Middleware())
	controllers.AuthRoutes(audited.Group("/auth"))
	controllers.InviteRoutes(audited.Group("/auth/invite"), controllers.NewInviteAuthClient(), controllers.NewInviteUsersClient())

	users := audited.Group("/users")
	users.Use(middleware.AuthRequired())
	controllers.UserRoutes(users)

	quotes := audited.Group("/quotes")
	quotes.Use(middleware.AuthRequired())
	controllers.QuotesRoutes(quotes, emailNotifier)

	exportGrp := audited.Group("/export")
	exportGrp.Use(middleware.AuthRequired())
	controllers.ExportRoutes(exportGrp)

	templates := audited.Group("/templates")
	templates.Use(middleware.AuthRequired())
	templates.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionTemplates))
	controllers.TemplateRoutes(templates)

	schedules := audited.Group("/schedules")
	schedules.Use(middleware.AuthRequired())
	schedules.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionSchedules))
	controllers.SchedulesRoutes(schedules, emailNotifier)

	fees := audited.Group("/fees")
	fees.Use(middleware.AuthRequired())
	fees.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionFees))
	controllers.FeesRoutes(fees)

	projects := audited.Group("/projects")
	projects.Use(middleware.AuthRequired())
	projects.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionSchedules))
	controllers.ProjectsRoutes(projects)

	invoices := audited.Group("/invoices")
	invoices.Use(middleware.AuthRequired())
	invoices.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	controllers.InvoicesRoutes(invoices)

	creditNotes := audited.Group("/credit-notes")
	creditNotes.Use(middleware.AuthRequired())
	creditNotes.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	controllers.CreditNotesRoutes(creditNotes)

	plans := audited.Group("/plans")
	plans.Use(middleware.AuthRequired())
	controllers.PlansRoutes(plans)

	subscriptions := audited.Group("/subscriptions")
	subscriptions.Use(middleware.AuthRequired())
	controllers.SubscriptionsRoutes(subscriptions)

	emailLogs := audited.Group("/email-logs")
	emailLogs.Use(middleware.AuthRequired())
	controllers.EmailLogsRoutes(emailLogs, authorizer)

	// logs group intentionally not under audited — audit reads must not log themselves
	logs := r.Group("/api/logs")
	logs.Use(middleware.AuthRequired())
	logs.Use(middleware.RequireSuperAdmin())
	controllers.AuditRoutes(logs, auditClient)

	return r
}
