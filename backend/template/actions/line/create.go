package line

import (
	"context"
	"database/sql"
	"strconv"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"
	templateGrpc "project-devis-template/services/grpc"

	"github.com/google/uuid"
)

func Create(ctx context.Context, db *sql.DB, req *templateGrpc.CreateTemplateLineRequest) (*templateGrpc.CreateTemplateLineResponse, error) {
	if req.Quantity != "" {
		if _, err := strconv.ParseFloat(req.Quantity, 64); err != nil {
			return &templateGrpc.CreateTemplateLineResponse{
				Success:          false,
				Code:             codes.InvalidInput,
				ValidationErrors: []*templateGrpc.ValidationError{sqlutil.Invalid("quantity", "Doit être un nombre valide.")},
			}, nil
		}
	}

	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM templates WHERE template_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		req.TemplateId, req.UserId,
	).Scan(&count)
	if err != nil || count == 0 {
		return &templateGrpc.CreateTemplateLineResponse{Success: false, Code: codes.NotFound}, nil
	}

	dataStr := req.Data
	if dataStr == "" {
		dataStr = "{}"
	}

	lineID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		`INSERT INTO template_lines (line_id, template_id, type, name, quantity, unit, unit_price, data, position, tax_id)
		 VALUES ($1, $2, $3, $4, $5::DECIMAL, $6, $7, $8::jsonb, $9, $10)`,
		lineID, req.TemplateId, req.Type, req.Name, req.Quantity,
		sqlutil.NullableStr(req.Unit), req.UnitPrice, dataStr, req.Position,
		sqlutil.NullableInt32(req.TaxId),
	)
	if err != nil {
		return &templateGrpc.CreateTemplateLineResponse{Success: false, Code: codes.InternalError}, nil
	}

	return &templateGrpc.CreateTemplateLineResponse{Success: true, Code: codes.Success, LineId: lineID}, nil
}
