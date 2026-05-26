package line

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *templateGrpc.DeleteTemplateLineRequest) (*templateGrpc.GenericResponse, error) {
	res, err := db.ExecContext(ctx,
		`DELETE FROM template_lines tl
		 USING templates t
		 WHERE tl.line_id=$1 AND tl.template_id=$2 AND t.template_id=$2 AND t.user_id=$3`,
		req.LineId, req.TemplateId, req.UserId,
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
