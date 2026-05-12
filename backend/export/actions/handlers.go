package actions

import (
	"context"

	quoteexport "project-devis-export/actions/quote"
	exportGrpc "project-devis-export/services/grpc"
)

func (s *Server) ExportQuote(ctx context.Context, req *exportGrpc.ExportQuoteRequest) (*exportGrpc.ExportQuoteResponse, error) {
	return quoteexport.Export(ctx, s.quote, s.users, s.gotenberg, req)
}
