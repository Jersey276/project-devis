package actions

import (
	"context"

	"project-devis-quote/actions/line"
	"project-devis-quote/actions/quote"
	quoteGrpc "project-devis-quote/services/grpc"
)

// ─── Quote ───────────────────────────────────────────────────────────────────

func (s *Server) CreateQuote(ctx context.Context, req *quoteGrpc.CreateQuoteRequest) (*quoteGrpc.CreateQuoteResponse, error) {
	return quote.Create(ctx, s.db, req)
}

func (s *Server) GetQuote(ctx context.Context, req *quoteGrpc.GetQuoteRequest) (*quoteGrpc.GetQuoteResponse, error) {
	return quote.Get(ctx, s.db, req)
}

func (s *Server) ListQuotes(ctx context.Context, req *quoteGrpc.ListQuotesRequest) (*quoteGrpc.ListQuotesResponse, error) {
	return quote.List(ctx, s.db, req)
}

func (s *Server) UpdateQuote(ctx context.Context, req *quoteGrpc.UpdateQuoteRequest) (*quoteGrpc.UpdateQuoteResponse, error) {
	return quote.Update(ctx, s.db, req)
}

func (s *Server) DeleteQuote(ctx context.Context, req *quoteGrpc.DeleteQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Delete(ctx, s.db, req)
}

func (s *Server) ArchiveQuote(ctx context.Context, req *quoteGrpc.ArchiveQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Archive(ctx, s.db, req)
}

func (s *Server) RestoreQuote(ctx context.Context, req *quoteGrpc.RestoreQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Restore(ctx, s.db, req)
}

func (s *Server) TrashQuotes(ctx context.Context, req *quoteGrpc.TrashQuotesRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Trash(ctx, s.db, req)
}

func (s *Server) DropQuote(ctx context.Context, req *quoteGrpc.DropQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Drop(ctx, s.db, req)
}

func (s *Server) ContinueQuote(ctx context.Context, req *quoteGrpc.ContinueQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	return quote.Continue(ctx, s.db, req)
}

// ─── Line ────────────────────────────────────────────────────────────────────

func (s *Server) CreateQuoteLine(ctx context.Context, req *quoteGrpc.CreateQuoteLineRequest) (*quoteGrpc.CreateQuoteLineResponse, error) {
	return line.Create(ctx, s.db, req)
}

func (s *Server) GetQuoteLine(ctx context.Context, req *quoteGrpc.GetQuoteLineRequest) (*quoteGrpc.GetQuoteLineResponse, error) {
	return line.Get(ctx, s.db, req)
}

func (s *Server) ListQuoteLines(ctx context.Context, req *quoteGrpc.ListQuoteLinesRequest) (*quoteGrpc.ListQuoteLinesResponse, error) {
	return line.List(ctx, s.db, req)
}

func (s *Server) UpdateQuoteLine(ctx context.Context, req *quoteGrpc.UpdateQuoteLineRequest) (*quoteGrpc.UpdateQuoteLineResponse, error) {
	return line.Update(ctx, s.db, req)
}

func (s *Server) DeleteQuoteLine(ctx context.Context, req *quoteGrpc.DeleteQuoteLineRequest) (*quoteGrpc.GenericResponse, error) {
	return line.Delete(ctx, s.db, req)
}
