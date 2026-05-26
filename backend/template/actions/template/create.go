package template

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	templateGrpc "project-devis-template/services/grpc"

	"github.com/google/uuid"
)

const (
	TypeQuoteDocument  = "quote_document"
	TypeQuoteLine      = "quote_line"
	TypeDocumentDesign = "document_design"
)

var validTypes = map[string]bool{
	TypeQuoteDocument:  true,
	TypeQuoteLine:      true,
	TypeDocumentDesign: true,
}

func Create(ctx context.Context, db *sql.DB, req *templateGrpc.CreateTemplateRequest) (*templateGrpc.CreateTemplateResponse, error) {
	if !validTypes[req.TemplateType] {
		return &templateGrpc.CreateTemplateResponse{Success: false, Code: codes.InvalidTemplateType}, nil
	}

	templateID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO templates (template_id, user_id, template_type, target_resource, name)
		 VALUES ($1, $2, $3, $4, $5)`,
		templateID, req.UserId, req.TemplateType, req.TargetResource, req.Name,
	)
	if err != nil {
		return &templateGrpc.CreateTemplateResponse{Success: false, Code: codes.InternalError}, nil
	}

	return &templateGrpc.CreateTemplateResponse{Success: true, Code: codes.Success, TemplateId: templateID}, nil
}
