package line

import (
	"context"
	"database/sql"
	"strconv"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"
	templateGrpc "project-devis-template/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *templateGrpc.UpdateTemplateLineRequest) (*templateGrpc.UpdateTemplateLineResponse, error) {
	if req.Quantity != "" {
		if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
			return &templateGrpc.UpdateTemplateLineResponse{
				Success:          false,
				Code:             codes.InvalidInput,
				ValidationErrors: []*templateGrpc.ValidationError{{Field: "quantity", Message: "Doit être un nombre valide."}},
			}, nil
		}
	}

	var dataStr interface{}
	if req.Data != "" {
		dataStr = req.Data
	}

	res, err := db.ExecContext(ctx,
		`UPDATE template_lines tl
		 SET type=COALESCE(NULLIF($1,''), tl.type),
		     name=COALESCE(NULLIF($2,''), tl.name),
		     quantity=COALESCE(NULLIF($3,'')::DECIMAL, tl.quantity),
		     unit=$4,
		     unit_price=$5,
		     data=CASE WHEN $6::text IS NOT NULL THEN $6::jsonb ELSE tl.data END,
		     position=$7,
		     tax_id=$8,
		     updated_at=now()
		 FROM templates t
		 WHERE tl.line_id=$9 AND tl.template_id=$10 AND t.template_id=$10 AND t.user_id=$11`,
		req.Type, req.Name, sqlutil.NullableStr(req.Quantity),
		sqlutil.NullableStr(req.Unit), req.UnitPrice,
		dataStr, req.Position, sqlutil.NullableInt32(req.TaxId),
		req.LineId, req.TemplateId, req.UserId,
	)
	if err != nil {
		return &templateGrpc.UpdateTemplateLineResponse{Success: false, Code: codes.InternalError}, nil
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return &templateGrpc.UpdateTemplateLineResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &templateGrpc.UpdateTemplateLineResponse{Success: true, Code: codes.Success}, nil
}
