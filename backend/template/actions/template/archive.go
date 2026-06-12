package template

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func Archive(ctx context.Context, db *sql.DB, req *templateGrpc.ArchiveTemplateRequest) (*templateGrpc.GenericResponse, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE templates SET archived_at=now(), updated_at=now()
		 WHERE template_id=$1 AND user_id=$2 AND archived_at IS NULL`,
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

func Restore(ctx context.Context, db *sql.DB, req *templateGrpc.RestoreTemplateRequest) (*templateGrpc.GenericResponse, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE templates SET archived_at=NULL, updated_at=now()
		 WHERE template_id=$1 AND user_id=$2 AND archived_at IS NOT NULL`,
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
