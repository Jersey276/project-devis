package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"gateway/audit"
	"gateway/middleware"

	"github.com/gin-gonic/gin"
)

var auditErrors = &serviceErrors{
	unavailableMessage: "Service de journal d'activité indisponible.",
	codes:              map[int32]codeMapping{},
}

func AuditRoutes(r *gin.RouterGroup, client audit.AuditServiceClient) {
	r.GET("", func(c *gin.Context) { listActivityLogs(c, client) })
	r.GET("/stats", func(c *gin.Context) { getActivityStats(c, client) })
	r.GET("/:id", func(c *gin.Context) { getActivityLog(c, client) })
	r.POST("/export", func(c *gin.Context) { exportActivityLogs(c, client) })
}

func parseRespStatuses(raw string) []int32 {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]int32, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.ParseInt(strings.TrimSpace(p), 10, 32)
		if err == nil && v > 0 {
			out = append(out, int32(v))
		}
	}
	return out
}

func listActivityLogs(c *gin.Context, client audit.AuditServiceClient) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "50"), 10, 32)

	resp, err := client.ListActivityLogs(c.Request.Context(), &audit.ListActivityLogsRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
		Filters: &audit.ActivityLogFilters{
			UserId:       c.Query("user_id"),
			UrlContains:  c.Query("url_contains"),
			RespStatuses: parseRespStatuses(c.Query("resp_statuses")),
			DateFrom:     c.Query("date_from"),
			DateTo:       c.Query("date_to"),
		},
	})
	if err != nil {
		auditErrors.unavailable(c)
		return
	}
	if !resp.Success {
		auditErrors.reply(c, resp.Code)
		return
	}

	logs := make([]gin.H, 0, len(resp.Logs))
	for _, l := range resp.Logs {
		logs = append(logs, gin.H{
			"id":          l.Id,
			"user_id":     l.UserId,
			"method":      l.Method,
			"url":         l.Url,
			"duration_ms": l.DurationMs,
			"resp_status": l.RespStatus,
			"created_at":  l.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"total":   resp.Total,
	})
}

func getActivityLog(c *gin.Context, client audit.AuditServiceClient) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "ID invalide."})
		return
	}

	resp, err := client.GetActivityLog(c.Request.Context(), &audit.GetActivityLogRequest{Id: id})
	if err != nil {
		auditErrors.unavailable(c)
		return
	}
	if !resp.Success {
		auditErrors.reply(c, resp.Code)
		return
	}

	l := resp.Log
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"log": gin.H{
			"id":          l.Id,
			"user_id":     l.UserId,
			"method":      l.Method,
			"url":         l.Url,
			"duration_ms": l.DurationMs,
			"req_body":    l.ReqBody,
			"resp_body":   l.RespBody,
			"resp_status": l.RespStatus,
			"created_at":  l.CreatedAt,
		},
	})
}

func getActivityStats(c *gin.Context, client audit.AuditServiceClient) {
	resp, err := client.GetActivityStats(c.Request.Context(), &audit.GetActivityStatsRequest{})
	if err != nil {
		auditErrors.unavailable(c)
		return
	}
	if !resp.Success {
		auditErrors.reply(c, resp.Code)
		return
	}

	stats := make([]gin.H, 0, len(resp.Stats))
	for _, s := range resp.Stats {
		stats = append(stats, gin.H{
			"date":        s.Date,
			"resp_status": s.RespStatus,
			"count":       s.Count,
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "stats": stats})
}

func exportActivityLogs(c *gin.Context, client audit.AuditServiceClient) {
	var body struct {
		Filters struct {
			UserID       string  `json:"user_id"`
			URLContains  string  `json:"url_contains"`
			RespStatuses []int32 `json:"resp_statuses"`
			DateFrom     string  `json:"date_from"`
			DateTo       string  `json:"date_to"`
		} `json:"filters"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Corps de requête invalide."})
		return
	}

	userID := c.GetString(middleware.CtxUserID)
	email := c.GetString(middleware.CtxEmail)

	resp, err := client.ExportActivityLogs(c.Request.Context(), &audit.ExportActivityLogsRequest{
		RecipientEmail: email,
		RecipientName:  userID,
		Filters: &audit.ActivityLogFilters{
			UserId:       body.Filters.UserID,
			UrlContains:  body.Filters.URLContains,
			RespStatuses: body.Filters.RespStatuses,
			DateFrom:     body.Filters.DateFrom,
			DateTo:       body.Filters.DateTo,
		},
	})
	if err != nil {
		auditErrors.unavailable(c)
		return
	}
	if !resp.Success {
		auditErrors.reply(c, resp.Code)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Export envoyé par email."})
}
