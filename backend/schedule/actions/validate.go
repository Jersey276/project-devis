package actions

import (
	"context"
	"database/sql"
	"strings"

	scheduleGrpc "project-devis-schedule/services/grpc"
)

func (s *Server) ValidateSchedule(ctx context.Context, req *scheduleGrpc.ValidateScheduleRequest) (*scheduleGrpc.GenericResponse, error) {
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

	var unbalancedCount int64
	err = tx.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM (
			SELECT sc.quote_line_id
			FROM schedule_cells sc
			JOIN quote_lines ql ON ql.line_id = sc.quote_line_id
			WHERE sc.schedule_id = $1
			GROUP BY sc.quote_line_id, ql.quantity, ql.unit_price
			HAVING COALESCE(SUM(sc.amount_cents), 0) <> COALESCE(ROUND(ql.unit_price * CAST(ql.quantity AS NUMERIC)), 0)
		) AS unbalanced_lines
	`, req.ScheduleId).Scan(&unbalancedCount)
	if err != nil {
		_ = tx.Rollback()
		return &scheduleGrpc.GenericResponse{Success: false, Code: CodeInternalError}, err
	}
	if unbalancedCount > 0 {
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