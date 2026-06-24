package controllers

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	quote "gateway/quote"
	schedule "gateway/schedule"
	gatewaySvc "gateway/services"
	users "gateway/users"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	ScheduleCodeNotFound      int32 = 1001
	ScheduleCodeAlreadyExists int32 = 1002
	ScheduleCodeInvalidInput  int32 = 1003
	ScheduleCodeFinalized     int32 = 1006
	ScheduleCodeUnbalanced    int32 = 1007
	ScheduleCodeValidated     int32 = 1008
	ScheduleCodeInternalError int32 = 2001
)

func scheduleValidationErrors(errs []*schedule.ValidationError) []FieldError {
	out := make([]FieldError, len(errs))
	for i, e := range errs {
		out[i] = FieldError{Field: e.Field, Message: e.Message}
	}
	return out
}

var scheduleErrors = &serviceErrors{
	codes: map[int32]codeMapping{
		ScheduleCodeNotFound:      {http.StatusNotFound, "Échéancier introuvable."},
		ScheduleCodeAlreadyExists: {http.StatusConflict, "Un échéancier existe déjà pour ce devis."},
		ScheduleCodeInvalidInput:  {http.StatusBadRequest, "Données invalides."},
		ScheduleCodeFinalized:     {http.StatusConflict, "Cet échéancier est finalisé et ne peut plus être modifié."},
		ScheduleCodeUnbalanced:    {http.StatusUnprocessableEntity, "L'échéancier n'est pas équilibré."},
		ScheduleCodeValidated:     {http.StatusConflict, "Cet échéancier est déjà validé."},
		ScheduleCodeInternalError: {http.StatusInternalServerError, "Une erreur interne est survenue."},
	},
	unavailableMessage: "Service échéancier indisponible.",
}

// SchedulesRoutes wires the /schedules API group against the schedule gRPC service.
func SchedulesRoutes(r *gin.RouterGroup, emailNotifier gatewaySvc.EmailNotifier) {
	address := os.Getenv("SCHEDULE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50056"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to schedule gRPC server: " + err.Error())
	}
	client := schedule.NewScheduleServiceClient(conn)

	quoteAddress := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if quoteAddress == "" {
		quoteAddress = "localhost:50053"
	}
	quoteConn, err := grpc.NewClient(quoteAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to quote gRPC server: " + err.Error())
	}
	quoteClient := quote.NewQuoteServiceClient(quoteConn)

	usersAddress := os.Getenv("USER_SERVICE_ADDRESS")
	if usersAddress == "" {
		usersAddress = "localhost:50052"
	}
	usersConn, err := grpc.NewClient(usersAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to users gRPC server: " + err.Error())
	}
	usersClient := users.NewUserServiceClient(usersConn)

	r.GET("", func(c *gin.Context) { ListSchedules(c, client) })
	r.POST("", func(c *gin.Context) { CreateSchedule(c, client, quoteClient) })

	one := r.Group("/:id")
	one.GET("", func(c *gin.Context) { GetSchedule(c, client) })
	one.PATCH("/cells", func(c *gin.Context) { UpdateScheduleCell(c, client) })
	one.PATCH("/status", func(c *gin.Context) { UpdateScheduleStatus(c, client, quoteClient, usersClient, emailNotifier) })
	one.POST("/validate", func(c *gin.Context) { ValidateSchedule(c, client, quoteClient, usersClient, emailNotifier) })
}

// ─── Email helper ────────────────────────────────────────────────────────────

func sendScheduleEmailNotification(
	scheduleID, userID, status string,
	scheduleClient schedule.ScheduleServiceClient,
	quoteClient quote.QuoteServiceClient,
	usersClient users.UserServiceClient,
	emailNotifier gatewaySvc.EmailNotifier,
) {
	ctx := context.Background()

	schedResp, err := scheduleClient.GetSchedule(ctx, &schedule.GetScheduleRequest{
		ScheduleId: scheduleID,
		UserId:     userID,
	})
	if err != nil || !schedResp.Success {
		log.Printf("schedule email: GetSchedule failed for %s: %v", scheduleID, err)
		return
	}

	quoteResp, err := quoteClient.GetQuote(ctx, &quote.GetQuoteRequest{
		QuoteId: schedResp.Schedule.QuoteId,
		UserId:  userID,
	})
	if err != nil || !quoteResp.Success {
		log.Printf("schedule email: GetQuote failed for %s: %v", schedResp.Schedule.QuoteId, err)
		return
	}

	clientResp, err := usersClient.GetClient(ctx, &users.GetClientRequest{
		ClientId: quoteResp.Quote.ClientId,
		UserId:   userID,
	})
	if err != nil || !clientResp.Success {
		log.Printf("schedule email: GetClient failed for %s: %v", quoteResp.Quote.ClientId, err)
		return
	}

	clientName := strings.TrimSpace(clientResp.Client.FirstName + " " + clientResp.Client.LastName)
	if err := emailNotifier.SendScheduleEmail(
		ctx, userID, quoteResp.Quote.QuoteId,
		clientResp.Client.Email, clientName, quoteResp.Quote.Name, status,
	); err != nil {
		log.Printf("schedule email: send failed for schedule %s: %v", scheduleID, err)
	}
}

// ─── Handlers ────────────────────────────────────────────────────────────────

func ListSchedules(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("list_schedules", success, grpcCode, startedAt)
	}()

	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var statuses []string
	if raw := c.Query("statuses"); raw != "" {
		statuses = strings.Split(raw, ",")
	}

	filterClientID := c.Query("client_id")
	// In customer mode, filter by client_id only — the authenticated user_id is
	// that of the client account, not the provider who owns the schedules.
	userID := userIDFromCtx(c)
	if filterClientID != "" && c.GetHeader("X-Client-Mode") == "customer" {
		userID = ""
	}

	resp, err := client.ListSchedules(c.Request.Context(), &schedule.ListSchedulesRequest{
		UserId:        userID,
		QuoteId:       c.Query("quote_id"),
		Page:          int32(page),
		PageSize:      int32(pageSize),
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortDirection: c.DefaultQuery("sort_direction", "desc"),
		Filters: &schedule.ScheduleFilters{
			Statuses:  statuses,
			StartFrom: c.Query("start_from"),
			StartTo:   c.Query("start_to"),
			ClientId:  filterClientID,
		},
	})
	if err != nil {
		grpcCode = ScheduleCodeInternalError
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		grpcCode = resp.Code
		scheduleErrors.reply(c, resp.Code)
		return
	}
	grpcCode = resp.Code
	success = true
	out := make([]gin.H, 0, len(resp.Schedules))
	for _, s := range resp.Schedules {
		out = append(out, gin.H{
			"schedule_id":     s.ScheduleId,
			"name":            s.Name,
			"status":          s.Status,
			"start_month":     s.StartMonth,
			"duration_months": s.DurationMonths,
			"quote_id":        s.QuoteId,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "schedules": out, "total": resp.Total})
}

func CreateSchedule(c *gin.Context, client schedule.ScheduleServiceClient, quoteClient quote.QuoteServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("create_schedule", success, grpcCode, startedAt)
	}()

	var input struct {
		QuoteID        string `json:"quote_id" binding:"required"`
		Name           string `json:"name" binding:"required"`
		StartMonth     string `json:"start_month" binding:"required"`
		DurationMonths int32  `json:"duration_months" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		grpcCode = ScheduleCodeInvalidInput
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	userID := userIDFromCtx(c)

	// Resolve client_id from the quote so schedules can be filtered by client.
	var clientID string
	if quoteResp, err := quoteClient.GetQuote(c.Request.Context(), &quote.GetQuoteRequest{
		QuoteId: input.QuoteID,
		UserId:  userID,
	}); err == nil && quoteResp.Success {
		clientID = quoteResp.Quote.GetClientId()
	}

	resp, err := client.CreateSchedule(c.Request.Context(), &schedule.CreateScheduleRequest{
		UserId:         userID,
		QuoteId:        input.QuoteID,
		Name:           input.Name,
		StartMonth:     input.StartMonth,
		DurationMonths: input.DurationMonths,
		ClientId:       clientID,
	})
	if err != nil {
		grpcCode = ScheduleCodeInternalError
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		grpcCode = resp.Code
		if len(resp.ValidationErrors) > 0 {
			scheduleErrors.replyWithValidation(c, resp.Code, scheduleValidationErrors(resp.ValidationErrors))
		} else {
			scheduleErrors.reply(c, resp.Code)
		}
		return
	}
	grpcCode = resp.Code
	success = true
	c.JSON(http.StatusCreated, gin.H{"success": true, "schedule_id": resp.ScheduleId})
}

func GetSchedule(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("get_schedule", success, grpcCode, startedAt)
	}()

	resp, err := client.GetSchedule(c.Request.Context(), &schedule.GetScheduleRequest{
		ScheduleId: c.Param("id"),
		UserId:     userIDFromCtx(c),
	})
	if err != nil {
		grpcCode = ScheduleCodeInternalError
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		grpcCode = resp.Code
		scheduleErrors.reply(c, resp.Code)
		return
	}
	grpcCode = resp.Code
	success = true
	s := resp.Schedule
	lines := make([]gin.H, 0, len(s.Lines))
	for _, l := range s.Lines {
		lines = append(lines, gin.H{
			"quote_line_id":  l.QuoteLineId,
			"planned_cents":  l.PlannedCents,
			"expected_cents": l.ExpectedCents,
		})
	}
	cols := make([]gin.H, 0, len(s.ColumnTotals))
	for _, ct := range s.ColumnTotals {
		cols = append(cols, gin.H{
			"month_index":  ct.MonthIndex,
			"amount_cents": ct.AmountCents,
		})
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"schedule": gin.H{
			"schedule_id":         s.ScheduleId,
			"quote_id":            s.QuoteId,
			"status":              s.Status,
			"name":                s.Name,
			"start_month":         s.StartMonth,
			"duration_months":     s.DurationMonths,
			"lines":               lines,
			"column_totals":       cols,
			"quote_total_cents":   s.QuoteTotalCents,
			"planned_total_cents": s.PlannedTotalCents,
		},
	})
}

func UpdateScheduleCell(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("update_schedule_cell", success, grpcCode, startedAt)
	}()

	var input struct {
		QuoteLineID string `json:"quote_line_id" binding:"required"`
		MonthIndex  int32  `json:"month_index" binding:"required"`
		AmountEur   string `json:"amount_eur" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		grpcCode = ScheduleCodeInvalidInput
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}
	resp, err := client.UpdateScheduleCell(c.Request.Context(), &schedule.UpdateScheduleCellRequest{
		ScheduleId:  c.Param("id"),
		UserId:      userIDFromCtx(c),
		QuoteLineId: input.QuoteLineID,
		MonthIndex:  input.MonthIndex,
		AmountEur:   input.AmountEur,
	})
	if err != nil {
		grpcCode = ScheduleCodeInternalError
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		grpcCode = resp.Code
		scheduleErrors.reply(c, resp.Code)
		return
	}
	grpcCode = resp.Code
	success = true
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func ValidateSchedule(
	c *gin.Context,
	client schedule.ScheduleServiceClient,
	quoteClient quote.QuoteServiceClient,
	usersClient users.UserServiceClient,
	emailNotifier gatewaySvc.EmailNotifier,
) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("validate_schedule", success, grpcCode, startedAt)
	}()

	scheduleID := c.Param("id")
	userID := userIDFromCtx(c)

	resp, err := client.ValidateSchedule(c.Request.Context(), &schedule.ValidateScheduleRequest{
		ScheduleId: scheduleID,
		UserId:     userID,
	})
	if err != nil {
		grpcCode = ScheduleCodeInternalError
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		grpcCode = resp.Code
		scheduleErrors.reply(c, resp.Code)
		return
	}
	grpcCode = resp.Code
	success = true

	go sendScheduleEmailNotification(scheduleID, userID, "VALID", client, quoteClient, usersClient, emailNotifier)

	c.JSON(http.StatusOK, gin.H{"success": true})
}
