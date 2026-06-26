package controllers

import (
	"net/http"
	"strconv"
	"strings"

	invoice "gateway/invoice"
	quote "gateway/quote"
	schedule "gateway/schedule"
	users "gateway/users"

	"github.com/gin-gonic/gin"
)

// GetAdminAccount fetches a single user account by user_id for admin use.
// It reuses ListAdminAccounts with a search on user_id and picks the first match.
func GetAdminAccount(c *gin.Context, client users.UserServiceClient) {
	userID := c.Param("userId")
	resp, err := client.ListAdminAccounts(c.Request.Context(), &users.ListAdminAccountsRequest{
		UserId: userID,
	})
	if err != nil {
		usersErrors.unavailable(c)
		return
	}
	if !resp.Success {
		usersErrors.reply(c, resp.Code)
		return
	}
	if len(resp.Accounts) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Utilisateur introuvable."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "user": marshalAdminAccount(resp.Accounts[0])})
}

// ListAdminUserQuotes returns the quotes owned by a specific user, for admin use.
func ListAdminUserQuotes(c *gin.Context, client quote.QuoteServiceClient) {
	targetUserID := c.Param("userId")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var states []string
	if raw := c.Query("states"); raw != "" {
		states = strings.Split(raw, ",")
	}

	resp, err := client.ListQuotes(c.Request.Context(), &quote.ListQuotesRequest{
		UserId:          targetUserID,
		IncludeArchived: c.Query("archived") == "true",
		Page:            int32(page),
		PageSize:        int32(pageSize),
		Filters: &quote.QuoteFilters{
			Search: c.Query("search"),
			States: states,
		},
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortDirection: c.DefaultQuery("sort_direction", "desc"),
	})
	if err != nil {
		quoteErrors.unavailable(c)
		return
	}
	if !resp.Success {
		quoteErrors.reply(c, resp.Code)
		return
	}

	out := make([]gin.H, 0, len(resp.Quotes))
	for _, q := range resp.Quotes {
		out = append(out, marshalQuote(q))
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "quotes": out, "total": resp.Total})
}

// ListAdminUserSchedules returns the schedules owned by a specific user, for admin use.
func ListAdminUserSchedules(c *gin.Context, client schedule.ScheduleServiceClient, quoteClient quote.QuoteServiceClient) {
	targetUserID := c.Param("userId")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var statuses []string
	if raw := c.Query("statuses"); raw != "" {
		statuses = strings.Split(raw, ",")
	}

	resp, err := client.ListSchedules(c.Request.Context(), &schedule.ListSchedulesRequest{
		UserId:        targetUserID,
		Page:          int32(page),
		PageSize:      int32(pageSize),
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortDirection: c.DefaultQuery("sort_direction", "desc"),
		Filters: &schedule.ScheduleFilters{
			Statuses: statuses,
		},
	})
	if err != nil {
		scheduleErrors.unavailable(c)
		return
	}
	if !resp.Success {
		scheduleErrors.reply(c, resp.Code)
		return
	}

	quoteNameByID := map[string]string{}
	quoteIDs := make([]string, 0, len(resp.Schedules))
	seen := map[string]struct{}{}
	for _, s := range resp.Schedules {
		if _, ok := seen[s.QuoteId]; !ok {
			seen[s.QuoteId] = struct{}{}
			quoteIDs = append(quoteIDs, s.QuoteId)
		}
	}
	if len(quoteIDs) > 0 {
		qResp, qErr := quoteClient.ListQuotes(c.Request.Context(), &quote.ListQuotesRequest{
			UserId:   targetUserID,
			Page:     1,
			PageSize: int32(len(quoteIDs)),
			Filters:  &quote.QuoteFilters{QuoteIds: quoteIDs},
		})
		if qErr == nil && qResp.GetSuccess() {
			for _, qt := range qResp.GetQuotes() {
				quoteNameByID[qt.GetQuoteId()] = qt.GetName()
			}
		}
	}

	out := make([]gin.H, 0, len(resp.Schedules))
	for _, s := range resp.Schedules {
		out = append(out, gin.H{
			"schedule_id":     s.ScheduleId,
			"name":            s.Name,
			"status":          s.Status,
			"start_month":     s.StartMonth,
			"duration_months": s.DurationMonths,
			"quote_id":        s.QuoteId,
			"quote_name":      quoteNameByID[s.QuoteId],
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "schedules": out, "total": resp.Total})
}

// ListAdminUserInvoices returns the invoices owned by a specific user, for admin use.
func ListAdminUserInvoices(c *gin.Context, client invoice.InvoiceServiceClient) {
	targetUserID := c.Param("userId")
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	pageSize, _ := strconv.ParseInt(c.DefaultQuery("page_size", "20"), 10, 32)

	var statuses []string
	if raw := c.Query("statuses"); raw != "" {
		statuses = strings.Split(raw, ",")
	}

	resp, err := client.ListInvoices(c.Request.Context(), &invoice.ListInvoicesRequest{
		UserId:        targetUserID,
		Page:          int32(page),
		PageSize:      int32(pageSize),
		SortBy:        c.DefaultQuery("sort_by", "created_at"),
		SortDirection: c.DefaultQuery("sort_direction", "desc"),
		Filters: &invoice.InvoiceFilters{
			Statuses: statuses,
		},
	})
	if err != nil {
		invoiceErrors.unavailable(c)
		return
	}
	if !resp.Success {
		invoiceErrors.reply(c, resp.Code)
		return
	}

	out := make([]gin.H, 0, len(resp.Invoices))
	for _, in := range resp.Invoices {
		out = append(out, gin.H{
			"invoice_id":      in.InvoiceId,
			"invoice_number":  in.InvoiceNumber,
			"status":          in.Status,
			"quote_id":        in.QuoteId,
			"schedule_id":     in.ScheduleId,
			"issued_at":       in.IssuedAt,
			"due_date":        in.DueDate,
			"total_ttc_cents": in.TotalTtcCents,
		})
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "invoices": out, "total": resp.Total})
}
