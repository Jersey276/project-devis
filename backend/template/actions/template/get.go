package template

import (
	"context"
	"database/sql"
	"encoding/json"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *templateGrpc.GetTemplateRequest) (*templateGrpc.GetTemplateResponse, error) {
	var t templateGrpc.Template
	var archivedAt sql.NullString
	var payloadRaw []byte

	err := db.QueryRowContext(ctx,
		`SELECT template_id, user_id, template_type, target_resource, name,
		        archived_at, payload_version, payload, created_at, updated_at
		 FROM templates WHERE template_id=$1 AND user_id=$2`,
		req.TemplateId, req.UserId,
	).Scan(
		&t.TemplateId, &t.UserId, &t.TemplateType, &t.TargetResource, &t.Name,
		&archivedAt, &t.PayloadVersion, &payloadRaw, &t.CreatedAt, &t.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return &templateGrpc.GetTemplateResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &templateGrpc.GetTemplateResponse{Success: false, Code: codes.InternalError}, nil
	}

	if archivedAt.Valid {
		t.ArchivedAt = archivedAt.String
	}
	t.Payload = string(json.RawMessage(payloadRaw))

	return &templateGrpc.GetTemplateResponse{Success: true, Code: codes.Success, Template: &t}, nil
}
