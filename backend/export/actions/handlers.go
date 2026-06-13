package actions

import (
	"context"

	invoiceexport "project-devis-export/actions/invoice"
	quoteexport "project-devis-export/actions/quote"
	scheduleexport "project-devis-export/actions/schedule"
	exportGrpc "project-devis-export/services/grpc"
)

func (s *Server) ExportQuote(ctx context.Context, req *exportGrpc.ExportQuoteRequest) (*exportGrpc.ExportQuoteResponse, error) {
	return quoteexport.Export(ctx, s.quote, s.users, s.gotenberg, req)
}

func (s *Server) ExportSchedule(ctx context.Context, req *exportGrpc.ExportScheduleRequest) (*exportGrpc.ExportQuoteResponse, error) {
	return scheduleexport.Export(ctx, s.schedule, s.quote, s.gotenberg, req)
}

func (s *Server) ExportInvoice(ctx context.Context, req *exportGrpc.ExportInvoiceRequest) (*exportGrpc.ExportQuoteResponse, error) {
	return invoiceexport.Export(ctx, s.invoice, s.gotenberg, req)
}
