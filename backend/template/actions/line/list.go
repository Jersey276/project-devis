package line

import (
	"context"
	"database/sql"
	"encoding/json"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *templateGrpc.ListTemplateLinesRequest) (*templateGrpc.ListTemplateLinesResponse, error) {
	var count int
	err := db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM templates WHERE template_id=$1 AND user_id=$2`,
		req.TemplateId, req.UserId,
	).Scan(&count)
	if err != nil || count == 0 {
		return &templateGrpc.ListTemplateLinesResponse{Success: false, Code: codes.NotFound}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT line_id, template_id, type, name, quantity::text, unit, unit_price, data, position, tax_id, created_at, updated_at
		 FROM template_lines WHERE template_id=$1 ORDER BY position ASC`,
		req.TemplateId,
	)
	if err != nil {
		return &templateGrpc.ListTemplateLinesResponse{Success: false, Code: codes.InternalError}, nil
	}
	defer rows.Close()

	lines := make([]*templateGrpc.TemplateLine, 0)
	for rows.Next() {
		var l templateGrpc.TemplateLine
		var unit sql.NullString
		var taxID sql.NullInt32
		var dataRaw []byte
		if err := rows.Scan(
			&l.LineId, &l.TemplateId, &l.Type, &l.Name, &l.Quantity,
			&unit, &l.UnitPrice, &dataRaw, &l.Position, &taxID,
			new(string), new(string),
		); err != nil {
			return &templateGrpc.ListTemplateLinesResponse{Success: false, Code: codes.InternalError}, nil
		}
		if unit.Valid {
			l.Unit = unit.String
		}
		if taxID.Valid {
			l.TaxId = taxID.Int32
		}
		l.Data = string(json.RawMessage(dataRaw))
		lines = append(lines, &l)
	}

	if err := rows.Err(); err != nil {
		return &templateGrpc.ListTemplateLinesResponse{Success: false, Code: codes.InternalError}, nil
	}

	return &templateGrpc.ListTemplateLinesResponse{Success: true, Code: codes.Success, Lines: lines}, nil
}
