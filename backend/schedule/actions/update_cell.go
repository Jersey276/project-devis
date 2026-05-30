package actions

import (
	"context"
	"database/sql"
	"strings"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) UpdateScheduleCell(ctx context.Context, req *scheduleGrpc.UpdateScheduleCellRequest) (*scheduleGrpc.GenericResponse, error) {
	if req == nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if strings.TrimSpace(req.ScheduleId) == "" || strings.TrimSpace(req.UserId) == "" || strings.TrimSpace(req.QuoteLineId) == "" {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}
	if req.MonthIndex <= 0 {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	amountCents, err := ParseAmountEuros(req.AmountEur)
	if err != nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	var status string
	err = s.db.QueryRowContext(ctx, `SELECT status FROM schedules WHERE schedule_id=$1 AND user_id=$2`, req.ScheduleId, req.UserId).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return &scheduleGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
		}
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if !IsEditableStatus(status) {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeScheduleFinalized}, nil
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE schedule_cells SET amount_cents=$1, updated_at=NOW() WHERE schedule_id=$2 AND quote_line_id=$3 AND month_index=$4`,
		amountCents, req.ScheduleId, req.QuoteLineId, req.MonthIndex,
	)
	if err != nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if affected == 0 {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}

	return &scheduleGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}