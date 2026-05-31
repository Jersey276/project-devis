package actions

import (
	"context"
	"strings"
	"time"

	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/google/uuid"
)

func (s *Server) CreateSchedule(ctx context.Context, req *scheduleGrpc.CreateScheduleRequest) (resp *scheduleGrpc.CreateScheduleResponse, err error) {
	eligibleLineIDs, err := fetchEligibleQuoteLineIDs(ctx, req)
	if err != nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInternalError}, err
	}
	return s.createScheduleWithEligibleLines(ctx, req, eligibleLineIDs)
}

// CreateScheduleWithEligibleLines is test-facing and allows bypassing quote RPC calls.
func (s *Server) CreateScheduleWithEligibleLines(ctx context.Context, req *scheduleGrpc.CreateScheduleRequest, eligibleLineIDs []string) (resp *scheduleGrpc.CreateScheduleResponse, err error) {
	return s.createScheduleWithEligibleLines(ctx, req, eligibleLineIDs)
}

func (s *Server) createScheduleWithEligibleLines(ctx context.Context, req *scheduleGrpc.CreateScheduleRequest, eligibleLineIDs []string) (resp *scheduleGrpc.CreateScheduleResponse, err error) {
	startedAt := time.Now()
	defer func() {
		code := CodeInternalError
		success := false
		if resp != nil {
			code = resp.Code
			success = resp.Success
		}
		recordOperation("create_schedule", success, code, startedAt, err)
	}()

	if req == nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if err := ValidateCreateScheduleInput(req.UserId, req.QuoteId, req.Name, req.StartMonth, req.DurationMonths); err != nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if len(eligibleLineIDs) == 0 {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	startMonthDate, err := parseStartMonth(req.StartMonth)
	if err != nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInternalError}, err
	}

	scheduleID := uuid.New().String()
	_, err = tx.ExecContext(ctx,
		`INSERT INTO schedules (schedule_id, quote_id, user_id, name, status, start_month, duration_months) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		scheduleID, req.QuoteId, req.UserId, req.Name, StatusDraft, startMonthDate, req.DurationMonths,
	)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInternalError}, err
	}

	for _, lineID := range eligibleLineIDs {
		trimmedLineID := strings.TrimSpace(lineID)
		if trimmedLineID == "" {
			_ = tx.Rollback()
			return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInvalidInput}, nil
		}
		for monthIndex := 1; monthIndex <= int(req.DurationMonths); monthIndex++ {
			_, err = tx.ExecContext(ctx,
				`INSERT INTO schedule_cells (schedule_id, quote_line_id, month_index, amount_cents) VALUES ($1, $2, $3, $4)`,
				scheduleID, trimmedLineID, monthIndex, int64(0),
			)
			if err != nil {
				_ = tx.Rollback()
				return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInternalError}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return &scheduleGrpc.CreateScheduleResponse{Success: false, Code: CodeInternalError}, err
	}

	return &scheduleGrpc.CreateScheduleResponse{Success: true, Code: CodeSuccess, ScheduleId: scheduleID}, nil
}

func fetchEligibleQuoteLineIDs(ctx context.Context, req *scheduleGrpc.CreateScheduleRequest) ([]string, error) {
	if req == nil {
		return nil, nil
	}

	amounts, err := getQuoteLineExpectedCents(ctx, req.UserId, req.QuoteId)
	if err != nil {
		return nil, err
	}

	lineIDs := make([]string, 0, len(amounts))
	for lineID := range amounts {
		lineIDs = append(lineIDs, lineID)
	}
	return lineIDs, nil
}

func parseStartMonth(input string) (time.Time, error) {
	return time.Parse("2006-01", input)
}
