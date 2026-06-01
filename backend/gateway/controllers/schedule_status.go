package controllers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	schedule "gateway/schedule"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	scheduleStatusDraft     = "DRAFT"
	scheduleStatusNegotiate = "NEGOCIATE"
	scheduleStatusDenied    = "DENIED"
	scheduleStatusValid     = "VALID"
)

func UpdateScheduleStatus(c *gin.Context, client schedule.ScheduleServiceClient) {
	startedAt := time.Now()
	grpcCode := int32(0)
	success := false
	defer func() {
		recordScheduleHTTP("update_schedule_status", success, grpcCode, startedAt)
	}()

	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		grpcCode = ScheduleCodeInvalidInput
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Données invalides."})
		return
	}

	status := strings.ToUpper(strings.TrimSpace(input.Status))
	switch status {
	case scheduleStatusValid:
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
		c.JSON(http.StatusOK, gin.H{"success": true, "status": status})
		return
	case scheduleStatusDraft, scheduleStatusNegotiate, scheduleStatusDenied:
		if err := updateScheduleStatusDirect(c.Request.Context(), c.Param("id"), userIDFromCtx(c), status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				grpcCode = ScheduleCodeNotFound
				scheduleErrors.reply(c, ScheduleCodeNotFound)
				return
			}
			grpcCode = ScheduleCodeInternalError
			scheduleErrors.unavailable(c)
			return
		}
		grpcCode = 0
		success = true
		c.JSON(http.StatusOK, gin.H{"success": true, "status": status})
		return
	default:
		grpcCode = ScheduleCodeInvalidInput
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Statut invalide."})
		return
	}
}

func updateScheduleStatusDirect(ctx context.Context, scheduleID, userID, status string) error {
	db, err := openScheduleDB()
	if err != nil {
		return err
	}
	defer db.Close()

	result, err := db.ExecContext(ctx,
		`UPDATE schedules SET status=$1, validated_at=NULL, updated_at=NOW() WHERE schedule_id=$2 AND user_id=$3`,
		status, scheduleID, userID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func openScheduleDB() (*sql.DB, error) {
	host := getenvDefault("DB_HOST", "postgres")
	port := getenvDefault("DB_PORT", "5432")
	user := getenvDefault("DB_USER", "devis-schedule")
	dbName := getenvDefault("DB_NAME", "schedule")
	passwordFile := getenvDefault("DB_PASSWORD_FILE", "/run/secrets/db_password")
	passwordBytes, err := os.ReadFile(passwordFile)
	if err != nil {
		return nil, fmt.Errorf("read db password: %w", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host,
		port,
		user,
		strings.TrimSpace(string(passwordBytes)),
		dbName,
	)
	return sql.Open("postgres", dsn)
}

func getenvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
