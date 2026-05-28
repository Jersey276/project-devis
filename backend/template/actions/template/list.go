package template

import (
	"context"
	"database/sql"
	"encoding/json"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *templateGrpc.ListTemplatesRequest) (*templateGrpc.ListTemplatesResponse, error) {
	query := `SELECT template_id, user_id, template_type, target_resource, name,
	                 archived_at, payload_version, payload, created_at, updated_at
	          FROM templates WHERE user_id=$1`
	args := []interface{}{req.UserId}

	if !req.IncludeArchived {
		query += " AND archived_at IS NULL"
	}
	if req.TemplateType != "" {
		args = append(args, req.TemplateType)
		query += " AND template_type=$2"
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &templateGrpc.ListTemplatesResponse{Success: false, Code: codes.InternalError}, nil
	}
	defer rows.Close()

	templates := make([]*templateGrpc.Template, 0)
	for rows.Next() {
		var t templateGrpc.Template
		var archivedAt sql.NullString
		var payloadRaw []byte
		if err := rows.Scan(
			&t.TemplateId, &t.UserId, &t.TemplateType, &t.TargetResource, &t.Name,
			&archivedAt, &t.PayloadVersion, &payloadRaw, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return &templateGrpc.ListTemplatesResponse{Success: false, Code: codes.InternalError}, nil
		}
		if archivedAt.Valid {
			t.ArchivedAt = archivedAt.String
		}
		t.Payload = string(json.RawMessage(payloadRaw))
		templates = append(templates, &t)
	}

	if err := rows.Err(); err != nil {
		return &templateGrpc.ListTemplatesResponse{Success: false, Code: codes.InternalError}, nil
	}

	return &templateGrpc.ListTemplatesResponse{Success: true, Code: codes.Success, Templates: templates}, nil
}
