package controllers

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	invoice "gateway/invoice"
	project "gateway/project"
	quote "gateway/quote"
	schedule "gateway/schedule"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ProjectCodeNotFound      int32 = 1001
	ProjectCodeAlreadyExists int32 = 1002
	ProjectCodeInvalidInput  int32 = 1003
	ProjectCodeInternalError int32 = 2001
)

func projectValidationErrors(errs []*project.ValidationError) []FieldError {
	out := make([]FieldError, len(errs))
	for i, e := range errs {
		out[i] = FieldError{Field: e.Field, Message: e.Message}
	}
	return out
}

var projectErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		ProjectCodeNotFound:      {http.StatusNotFound, "Projet introuvable."},
		ProjectCodeAlreadyExists: {http.StatusConflict, "Ce projet existe déjà."},
		ProjectCodeInvalidInput:  {http.StatusBadRequest, "Données invalides."},
		ProjectCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service projet indisponible.",
}

// ProjectsRoutes wires the /projects API group against the project gRPC service.
func ProjectsRoutes(r *gin.RouterGroup) {
	addr := os.Getenv("PROJECT_SERVICE_ADDRESS")
	if addr == "" {
		addr = "localhost:50061"
	}
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("project: failed to connect to project service: " + err.Error())
	}
	client := project.NewProjectServiceClient(conn)

	// Downstream clients for the aggregated detail endpoint
	quoteAddr := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if quoteAddr == "" {
		quoteAddr = "localhost:50053"
	}
	quoteConn, err := grpc.NewClient(quoteAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("project: failed to connect to quote service: " + err.Error())
	}
	quoteClient := quote.NewQuoteServiceClient(quoteConn)

	scheduleAddr := os.Getenv("SCHEDULE_SERVICE_ADDRESS")
	if scheduleAddr == "" {
		scheduleAddr = "localhost:50056"
	}
	scheduleConn, err := grpc.NewClient(scheduleAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("project: failed to connect to schedule service: " + err.Error())
	}
	scheduleClient := schedule.NewScheduleServiceClient(scheduleConn)

	invoiceAddr := os.Getenv("INVOICE_SERVICE_ADDRESS")
	if invoiceAddr == "" {
		invoiceAddr = "localhost:50059"
	}
	invoiceConn, err := grpc.NewClient(invoiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("project: failed to connect to invoice service: " + err.Error())
	}
	invoiceClient := invoice.NewInvoiceServiceClient(invoiceConn)

	r.GET("", func(c *gin.Context) { listProjects(c, client) })
	r.POST("", func(c *gin.Context) { createProject(c, client) })
	r.GET("/:id", func(c *gin.Context) { getProject(c, client) })
	r.PUT("/:id", func(c *gin.Context) { updateProject(c, client) })
	r.DELETE("/:id", func(c *gin.Context) { deleteProject(c, client) })
	r.POST("/:id/quotes", func(c *gin.Context) { addQuoteToProject(c, client) })
	r.DELETE("/:id/quotes/:quoteId", func(c *gin.Context) { removeQuoteFromProject(c, client) })
	r.GET("/:id/detail", func(c *gin.Context) {
		getProjectDetail(c, client, quoteClient, scheduleClient, invoiceClient)
	})
}

func listProjects(c *gin.Context, client project.ProjectServiceClient) {
	userID := userIDFromCtx(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.ListProjects(ctx, &project.ListProjectsRequest{
		UserId:        userID,
		Page:          int32(page),
		PageSize:      int32(pageSize),
		Search:        c.Query("search"),
		Status:        c.Query("status"),
		ClientId:      c.Query("client_id"),
		SortBy:        c.Query("sort_by"),
		SortDirection: c.Query("sort_direction"),
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "projects": resp.Projects, "total": resp.Total})
}

func createProject(c *gin.Context, client project.ProjectServiceClient) {
	var body struct {
		Name     string `json:"name"`
		ClientId string `json:"client_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Corps de requête invalide."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.CreateProject(ctx, &project.CreateProjectRequest{
		UserId:   userIDFromCtx(c),
		Name:     body.Name,
		ClientId: body.ClientId,
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.replyWithValidation(c, resp.Code, projectValidationErrors(resp.ValidationErrors))
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "project_id": resp.ProjectId})
}

func getProject(c *gin.Context, client project.ProjectServiceClient) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.GetProject(ctx, &project.GetProjectRequest{
		ProjectId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "project": resp.Project})
}

func updateProject(c *gin.Context, client project.ProjectServiceClient) {
	var body struct {
		Name     string `json:"name"`
		ClientId string `json:"client_id"`
		Status   string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Corps de requête invalide."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.UpdateProject(ctx, &project.UpdateProjectRequest{
		ProjectId: c.Param("id"),
		UserId:    userIDFromCtx(c),
		Name:      body.Name,
		ClientId:  body.ClientId,
		Status:    body.Status,
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func deleteProject(c *gin.Context, client project.ProjectServiceClient) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.DeleteProject(ctx, &project.DeleteProjectRequest{
		ProjectId: c.Param("id"),
		UserId:    userIDFromCtx(c),
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func addQuoteToProject(c *gin.Context, client project.ProjectServiceClient) {
	var body struct {
		QuoteId string `json:"quote_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Corps de requête invalide."})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.AddQuoteToProject(ctx, &project.AddQuoteToProjectRequest{
		ProjectId: c.Param("id"),
		UserId:    userIDFromCtx(c),
		QuoteId:   body.QuoteId,
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func removeQuoteFromProject(c *gin.Context, client project.ProjectServiceClient) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	resp, err := client.RemoveQuoteFromProject(ctx, &project.RemoveQuoteFromProjectRequest{
		ProjectId: c.Param("id"),
		UserId:    userIDFromCtx(c),
		QuoteId:   c.Param("quoteId"),
	})
	if err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !resp.Success {
		projectErrors.reply(c, resp.Code)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// projectQuoteRow is the shape returned per quote in the detail response.
type projectQuoteRow struct {
	QuoteId   string                      `json:"quote_id"`
	UserId    string                      `json:"user_id"`
	Name      string                      `json:"name"`
	State     string                      `json:"state"`
	ClientId  string                      `json:"client_id"`
	Archived  bool                        `json:"archived"`
	Schedules []*schedule.ScheduleSummary `json:"schedules"`
	Invoices  []*invoice.InvoiceSummary   `json:"invoices"`
}

func getProjectDetail(
	c *gin.Context,
	projectClient project.ProjectServiceClient,
	quoteClient quote.QuoteServiceClient,
	scheduleClient schedule.ScheduleServiceClient,
	invoiceClient invoice.InvoiceServiceClient,
) {
	projectID := c.Param("id")
	userID := userIDFromCtx(c)

	ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()

	// Step 1: fetch project metadata and its quote IDs in parallel
	var projectResp *project.GetProjectResponse
	var quoteIdsResp *project.ListProjectQuoteIdsResponse

	eg1, ctx1 := errgroup.WithContext(ctx)
	eg1.Go(func() error {
		var err error
		projectResp, err = projectClient.GetProject(ctx1, &project.GetProjectRequest{
			ProjectId: projectID,
			UserId:    userID,
		})
		return err
	})
	eg1.Go(func() error {
		var err error
		quoteIdsResp, err = projectClient.ListProjectQuoteIds(ctx1, &project.ListProjectQuoteIdsRequest{
			ProjectId: projectID,
			UserId:    userID,
		})
		return err
	})

	if err := eg1.Wait(); err != nil {
		projectErrors.unavailable(c)
		return
	}
	if !projectResp.Success {
		projectErrors.reply(c, projectResp.Code)
		return
	}

	quoteIDs := quoteIdsResp.QuoteIds

	// Step 2: fan-out to quote/schedule/invoice services
	var quotes []*quote.Quote
	var schedules []*schedule.ScheduleSummary
	var invoices []*invoice.InvoiceSummary

	eg2, ctx2 := errgroup.WithContext(ctx)

	if len(quoteIDs) > 0 {
		eg2.Go(func() error {
			resp, err := quoteClient.ListQuotes(ctx2, &quote.ListQuotesRequest{
				UserId:   userID,
				PageSize: 200,
				Filters:  &quote.QuoteFilters{QuoteIds: quoteIDs},
			})
			if err != nil {
				return err
			}
			quotes = resp.Quotes
			return nil
		})
		eg2.Go(func() error {
			resp, err := scheduleClient.ListSchedules(ctx2, &schedule.ListSchedulesRequest{
				UserId:   userID,
				PageSize: 200,
				QuoteIds: quoteIDs,
			})
			if err != nil {
				return err
			}
			schedules = resp.Schedules
			return nil
		})
		eg2.Go(func() error {
			resp, err := invoiceClient.ListInvoices(ctx2, &invoice.ListInvoicesRequest{
				UserId:   userID,
				PageSize: 200,
				Filters:  &invoice.InvoiceFilters{QuoteIds: quoteIDs},
			})
			if err != nil {
				return err
			}
			invoices = resp.Invoices
			return nil
		})
	}

	if err := eg2.Wait(); err != nil {
		projectErrors.unavailable(c)
		return
	}

	// Group schedules and invoices by quote_id
	schedulesByQuote := make(map[string][]*schedule.ScheduleSummary)
	for _, s := range schedules {
		schedulesByQuote[s.QuoteId] = append(schedulesByQuote[s.QuoteId], s)
	}
	invoicesByQuote := make(map[string][]*invoice.InvoiceSummary)
	for _, inv := range invoices {
		invoicesByQuote[inv.QuoteId] = append(invoicesByQuote[inv.QuoteId], inv)
	}

	// Build quote rows preserving project order
	quoteMap := make(map[string]*quote.Quote, len(quotes))
	for _, q := range quotes {
		quoteMap[q.QuoteId] = q
	}

	rows := make([]projectQuoteRow, 0, len(quoteIDs))
	for _, qid := range quoteIDs {
		q := quoteMap[qid]
		if q == nil {
			continue
		}
		state := stateToLower(q.State)
		rows = append(rows, projectQuoteRow{
			QuoteId:   q.QuoteId,
			UserId:    q.UserId,
			Name:      q.Name,
			State:     state,
			ClientId:  q.ClientId,
			Archived:  q.Archived,
			Schedules: schedulesByQuote[qid],
			Invoices:  invoicesByQuote[qid],
		})
	}

	// Compute revenue totals from invoices
	var totalHT, collectedHT int64
	for _, inv := range invoices {
		if inv.Status == "ISSUED" || inv.Status == "PAID" {
			totalHT += inv.TotalHtCents
		}
		if inv.Status == "PAID" {
			collectedHT += inv.TotalHtCents
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":            true,
		"project":            projectResp.Project,
		"quotes":             rows,
		"total_ht_cents":     totalHT,
		"collected_ht_cents": collectedHT,
	})
}
