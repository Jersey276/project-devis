package controllers

import (
	"net/http"
	"os"
	"time"

	schedule "gateway/schedule"

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
func SchedulesRoutes(r *gin.RouterGroup) {
	address := os.Getenv("SCHEDULE_SERVICE_ADDRESS")
	if address == "" {
		address = "localhost:50056"
	}
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic("failed to connect to schedule gRPC server: " + err.Error())
	}
	client := schedule.NewScheduleServiceClient(conn)

	r.GET("", func(c *gin.Context) { ListSchedules(c, client) })
	r.POST("", func(c *gin.Context) { CreateSchedule(c, client) })

	one := r.Group("/:id")
	one.GET("", func(c *gin.Context) { GetSchedule(c, client) })
	one.PATCH("/cells", func(c *gin.Context) { UpdateScheduleCell(c, client) })
	one.POST("/validate", func(c *gin.Context) { ValidateSchedule(c, client) })
}

// ─── Handlers ────────────────────────────────────────────────────────────────

func ListSchedules(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("list_schedules", success, grpcCode, startedAt)
	}()

	resp, err := client.ListSchedules(c.Request.Context(), &schedule.ListSchedulesRequest{
		UserId:  userIDFromCtx(c),
		QuoteId: c.Query("quote_id"),
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
	c.JSON(http.StatusOK, gin.H{"success": true, "schedules": out})
}

func CreateSchedule(c *gin.Context, client schedule.ScheduleServiceClient) {
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
	resp, err := client.CreateSchedule(c.Request.Context(), &schedule.CreateScheduleRequest{
		UserId:         userIDFromCtx(c),
		QuoteId:        input.QuoteID,
		Name:           input.Name,
		StartMonth:     input.StartMonth,
		DurationMonths: input.DurationMonths,
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

func ValidateSchedule(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("validate_schedule", success, grpcCode, startedAt)
	}()

	resp, err := client.ValidateSchedule(c.Request.Context(), &schedule.ValidateScheduleRequest{
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
	c.JSON(http.StatusOK, gin.H{"success": true})
}
