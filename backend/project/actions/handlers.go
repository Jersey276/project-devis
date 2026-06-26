package actions

import (
	"context"

	"project-devis-project/actions/project"
	projectGrpc "project-devis-project/services/grpc"
)

func (s *Server) CreateProject(ctx context.Context, req *projectGrpc.CreateProjectRequest) (*projectGrpc.CreateProjectResponse, error) {
	return project.Create(ctx, s.db, req)
}

func (s *Server) GetProject(ctx context.Context, req *projectGrpc.GetProjectRequest) (*projectGrpc.GetProjectResponse, error) {
	return project.Get(ctx, s.db, req)
}

func (s *Server) ListProjects(ctx context.Context, req *projectGrpc.ListProjectsRequest) (*projectGrpc.ListProjectsResponse, error) {
	return project.List(ctx, s.db, req)
}

func (s *Server) UpdateProject(ctx context.Context, req *projectGrpc.UpdateProjectRequest) (*projectGrpc.GenericResponse, error) {
	return project.Update(ctx, s.db, req)
}

func (s *Server) DeleteProject(ctx context.Context, req *projectGrpc.DeleteProjectRequest) (*projectGrpc.GenericResponse, error) {
	return project.Delete(ctx, s.db, req)
}

func (s *Server) AddQuoteToProject(ctx context.Context, req *projectGrpc.AddQuoteToProjectRequest) (*projectGrpc.GenericResponse, error) {
	return project.AddQuote(ctx, s.db, req)
}

func (s *Server) RemoveQuoteFromProject(ctx context.Context, req *projectGrpc.RemoveQuoteFromProjectRequest) (*projectGrpc.GenericResponse, error) {
	return project.RemoveQuote(ctx, s.db, req)
}

func (s *Server) ListProjectQuoteIds(ctx context.Context, req *projectGrpc.ListProjectQuoteIdsRequest) (*projectGrpc.ListProjectQuoteIdsResponse, error) {
	return project.ListQuoteIds(ctx, s.db, req)
}
