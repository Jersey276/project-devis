package actions

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
)

func (s *Server) CreateInvoiceFromQuote(ctx context.Context, req *invoiceGrpc.CreateInvoiceFromQuoteRequest) (resp *invoiceGrpc.CreateInvoiceResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("create_invoice_from_quote", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var fieldErrors []*invoiceGrpc.ValidationError
	if strings.TrimSpace(req.UserId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if strings.TrimSpace(req.QuoteId) == "" {
		fieldErrors = append(fieldErrors, &invoiceGrpc.ValidationError{Field: "quote_id", Message: "Champ requis."})
	}
	if len(fieldErrors) > 0 {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	quoteResp, err := s.quoteClient.GetQuote(ctx, &quoteGrpc.GetQuoteRequest{QuoteId: req.QuoteId, UserId: req.UserId})
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	if !quoteResp.GetSuccess() || quoteResp.GetQuote() == nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.NotFound}, nil
	}
	if quoteResp.GetQuote().GetState() != quoteGrpc.QuoteState_QUOTE_STATE_VALIDATED {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.SourceNotEligible}, nil
	}

	listResp, err := s.scheduleClient.ListSchedules(ctx, &scheduleGrpc.ListSchedulesRequest{
		UserId:  req.UserId,
		QuoteId: req.QuoteId,
	})
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	if listResp.GetSuccess() && len(listResp.GetSchedules()) > 0 {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.QuoteHasSchedule}, nil
	}

	invoiceID := uuid.New().String()
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO invoices (invoice_id, user_id, quote_id, status)
		 VALUES ($1, $2, $3, 'DRAFT')`,
		invoiceID, req.UserId, req.QuoteId,
	)
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	if req.IssueNow {
		return s.issue(ctx, invoiceID, req.UserId, req.SaleDate, req.DueInDays)
	}
	return &invoiceGrpc.CreateInvoiceResponse{Success: true, Code: codes.Success, InvoiceId: invoiceID}, nil
}
