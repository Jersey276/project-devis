package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) ValidateSchedule(ctx context.Context, req *scheduleGrpc.ValidateScheduleRequest) (resp *scheduleGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer func() {
		code := CodeInternalError
		success := false
		if resp != nil {
			code = resp.Code
			success = resp.Success
		}
		recordOperation("validate_schedule", success, code, startedAt, err)
	}()

	if req == nil || strings.TrimSpace(req.ScheduleId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	var quoteID, status string
	err = tx.QueryRowContext(ctx, `SELECT quote_id, status FROM schedules WHERE schedule_id=$1 AND user_id=$2 FOR UPDATE`, req.ScheduleId, req.UserId).Scan(&quoteID, &status)
	if err != nil {
		_ = tx.Rollback()
		if err == sql.ErrNoRows {
			return &scheduleGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
		}
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if status == StatusValid {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeScheduleValidated}, nil
	}
	if status == StatusDenied {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeScheduleFinalized}, nil
	}

	var existingValidCount int64
	err = tx.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schedules WHERE quote_id=$1 AND schedule_id<>$2 AND status='VALID'`,
		quoteID, req.ScheduleId,
	).Scan(&existingValidCount)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if existingValidCount > 0 {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeScheduleValidated}, nil
	}

	expectedByLineID, err := getQuoteLineExpectedCents(ctx, req.UserId, quoteID)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	plannedByLineID := make(map[string]int64)
	rows, err := tx.QueryContext(ctx,
		`SELECT quote_line_id, COALESCE(SUM(amount_cents), 0) FROM schedule_cells WHERE schedule_id=$1 GROUP BY quote_line_id`,
		req.ScheduleId,
	)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	defer rows.Close()

	for rows.Next() {
		var lineID string
		var planned int64
		if err := rows.Scan(&lineID, &planned); err != nil {
			_ = tx.Rollback()
			return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
		}
		plannedByLineID[lineID] = planned
	}
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	unbalanced := false
	for lineID, expected := range expectedByLineID {
		if plannedByLineID[lineID] != expected {
			unbalanced = true
			break
		}
	}
	if !unbalanced {
		for lineID, planned := range plannedByLineID {
			if _, ok := expectedByLineID[lineID]; !ok && planned != 0 {
				unbalanced = true
				break
			}
		}
	}

	if unbalanced {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeScheduleUnbalanced}, nil
	}

	result, err := tx.ExecContext(ctx,
		`UPDATE schedules SET status='VALID', validated_at=NOW(), updated_at=NOW() WHERE schedule_id=$1 AND user_id=$2`,
		req.ScheduleId, req.UserId,
	)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if affected == 0 {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeNotFound}, nil
	}

	_, err = tx.ExecContext(ctx,
		`UPDATE schedules SET status='DENIED', updated_at=NOW() WHERE quote_id=$1 AND schedule_id<>$2 AND status IN ('DRAFT','NEGOCIATE')`,
		quoteID, req.ScheduleId,
	)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	if err := tx.Commit(); err != nil {
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}

	return &scheduleGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
