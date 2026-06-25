package template

import (
	"context"
	"database/sql"

	"project-devis-template/actions/codes"
	"project-devis-template/actions/sqlutil"
	templateGrpc "project-devis-template/services/grpc"

	"github.com/google/uuid"
)

func Create(ctx context.Context, db *sql.DB, req *templateGrpc.CreateTemplateRequest) (*templateGrpc.CreateTemplateResponse, error) {
	var fieldErrors []*templateGrpc.ValidationError

	if req.UserId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("user_id"))
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}
	if !sqlutil.ValidateTemplateType(req.TemplateType) {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("template_type", "Type de template invalide."))
	}

	if len(fieldErrors) > 0 {
		code := codes.InvalidInput
		if !sqlutil.ValidateTemplateType(req.TemplateType) && req.UserId != "" && req.Name != "" {
			code = codes.InvalidTemplateType
		}
		return &templateGrpc.CreateTemplateResponse{Success: false, Code: code, ValidationErrors: fieldErrors}, nil
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
