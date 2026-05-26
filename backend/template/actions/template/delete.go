package template

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *templateGrpc.DeleteTemplateRequest) (*templateGrpc.GenericResponse, error) {
	res, err := db.ExecContext(ctx,
		`DELETE FROM templates WHERE template_id=$1 AND user_id=$2`,
		req.TemplateId, req.UserId,
	)
	if err != nil {
		return &templateGrpc.GenericResponse{Success: false, Code: codes.InternalError}, nil
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &templateGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &templateGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
