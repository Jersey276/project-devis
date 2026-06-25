package main

import (
	"os"

	"gateway/audit"
	"gateway/authz"
	"gateway/controllers"
	"gateway/middleware"
	"gateway/services"
	quote "gateway/quote"
	users "gateway/users"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func newUsersClient() users.UserServiceClient {
	address := os.Getenv("USER_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50052"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to users gRPC server: " + err.Error())
	}
	return users.NewUserServiceClient(conn)
}

func setupRouter(auditLogger *middleware.AuditLogger, auditClient audit.AuditServiceClient) *gin.Engine {
	r := gin.Default()

	emailNotifier := services.NewEmailNotifier()
	usersClient := newUsersClient()

	// Webhooks — no auth, raw body, registered before auth groups
	webhooks := r.Group("/api/webhooks")
	controllers.WebhookRoutes(webhooks)
	controllers.ResendWebhookRoutes(webhooks)

	audited := r.Group("/api")
	audited.Use(auditLogger.Middleware())
	controllers.AuthRoutes(audited.Group("/auth"))
	controllers.InviteRoutes(audited.Group("/auth/invite"), controllers.NewInviteAuthClient(), controllers.NewInviteUsersClient())

	usersGroup := audited.Group("/users")
	usersGroup.Use(middleware.AuthRequired())
	controllers.UserRoutes(usersGroup)

	quotes := audited.Group("/quotes")
	quotes.Use(middleware.AuthRequired())
	controllers.QuotesRoutes(quotes, emailNotifier)

	exportGrp := audited.Group("/export")
	exportGrp.Use(middleware.AuthRequired())
	controllers.ExportRoutes(exportGrp, usersClient)

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

	quoteAddress := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if quoteAddress == "" {
		quoteAddress = "localhost:50053"
	}
	quoteConn, err := grpc.NewClient(quoteAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to quote gRPC server: " + err.Error())
	}
	quoteClient := quote.NewQuoteServiceClient(quoteConn)

	invoices := audited.Group("/invoices")
	invoices.Use(middleware.AuthRequired())
	invoices.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	controllers.InvoicesRoutes(invoices, usersClient, quoteClient)

	creditNotes := audited.Group("/credit-notes")
	creditNotes.Use(middleware.AuthRequired())
	creditNotes.Use(middleware.RequireSubscriptionFeature(authz.ResourceSubscriptionInvoices))
	creditNotes.Use(controllers.DenyCustomer())
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
