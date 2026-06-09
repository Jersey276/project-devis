package actions

import (
	"context"
	"log"

	emailGrpc "project-devis-email/services/grpc"
)

func (s *Server) TrackEmailEvent(ctx context.Context, req *emailGrpc.TrackEmailEventRequest) (*emailGrpc.GenericResponse, error) {
	if req.ResendId == "" {
		return &emailGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE email_logs
		 SET opened_at  = CASE WHEN $2 = 'email.opened'  AND opened_at  IS NULL THEN $3::timestamp ELSE opened_at  END,
		     clicked_at = CASE WHEN $2 = 'email.clicked' AND clicked_at IS NULL THEN $3::timestamp ELSE clicked_at END
		 WHERE resend_id = $1`,
		req.ResendId, req.EventType, req.OccurredAt,
	)
	if err != nil {
		log.Printf("track email event failed resend_id=%s event=%s: %v", req.ResendId, req.EventType, err)
		return &emailGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	return &emailGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
