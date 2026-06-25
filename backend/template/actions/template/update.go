package template

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"
	templateGrpc "project-devis-template/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *templateGrpc.UpdateTemplateRequest) (*templateGrpc.UpdateTemplateResponse, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE templates SET name=COALESCE(NULLIF($1,''), name),
		        target_resource=COALESCE(NULLIF($2,''), target_resource),
		        payload=CASE WHEN $3::text IS NOT NULL AND $3::text != '' THEN $3::jsonb ELSE payload END,
		        payload_version=CASE WHEN $4 > 0 THEN $4 ELSE payload_version END,
		        updated_at=now()
		 WHERE template_id=$5 AND user_id=$6 AND archived_at IS NULL`,
		req.Name, req.TargetResource, sqlutil.NullableStr(req.Payload), req.PayloadVersion,
		req.TemplateId, req.UserId,
	)
	if err != nil {
		return &templateGrpc.UpdateTemplateResponse{Success: false, Code: codes.InternalError}, nil
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &templateGrpc.UpdateTemplateResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &templateGrpc.UpdateTemplateResponse{Success: true, Code: codes.Success}, nil
}

