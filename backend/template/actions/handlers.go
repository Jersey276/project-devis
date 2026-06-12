package actions

import (
	"context"

	"project-devis-template/actions/line"
	templateAction "project-devis-template/actions/template"
	templateGrpc "project-devis-template/services/grpc"
)

// ─── Template ─────────────────────────────────────────────────────────────────

func (s *Server) ListTemplates(ctx context.Context, req *templateGrpc.ListTemplatesRequest) (*templateGrpc.ListTemplatesResponse, error) {
	return templateAction.List(ctx, s.db, req)
}

func (s *Server) CreateTemplate(ctx context.Context, req *templateGrpc.CreateTemplateRequest) (*templateGrpc.CreateTemplateResponse, error) {
	return templateAction.Create(ctx, s.db, req)
}

func (s *Server) GetTemplate(ctx context.Context, req *templateGrpc.GetTemplateRequest) (*templateGrpc.GetTemplateResponse, error) {
	return templateAction.Get(ctx, s.db, req)
}

func (s *Server) UpdateTemplate(ctx context.Context, req *templateGrpc.UpdateTemplateRequest) (*templateGrpc.UpdateTemplateResponse, error) {
	return templateAction.Update(ctx, s.db, req)
}

func (s *Server) DeleteTemplate(ctx context.Context, req *templateGrpc.DeleteTemplateRequest) (*templateGrpc.GenericResponse, error) {
	return templateAction.Delete(ctx, s.db, req)
}

func (s *Server) ArchiveTemplate(ctx context.Context, req *templateGrpc.ArchiveTemplateRequest) (*templateGrpc.GenericResponse, error) {
	return templateAction.Archive(ctx, s.db, req)
}

func (s *Server) RestoreTemplate(ctx context.Context, req *templateGrpc.RestoreTemplateRequest) (*templateGrpc.GenericResponse, error) {
	return templateAction.Restore(ctx, s.db, req)
}

// ─── Line ─────────────────────────────────────────────────────────────────────

func (s *Server) ListTemplateLines(ctx context.Context, req *templateGrpc.ListTemplateLinesRequest) (*templateGrpc.ListTemplateLinesResponse, error) {
	return line.List(ctx, s.db, req)
}

func (s *Server) CreateTemplateLine(ctx context.Context, req *templateGrpc.CreateTemplateLineRequest) (*templateGrpc.CreateTemplateLineResponse, error) {
	return line.Create(ctx, s.db, req)
}

func (s *Server) UpdateTemplateLine(ctx context.Context, req *templateGrpc.UpdateTemplateLineRequest) (*templateGrpc.UpdateTemplateLineResponse, error) {
	return line.Update(ctx, s.db, req)
}

func (s *Server) DeleteTemplateLine(ctx context.Context, req *templateGrpc.DeleteTemplateLineRequest) (*templateGrpc.GenericResponse, error) {
	return line.Delete(ctx, s.db, req)
}
