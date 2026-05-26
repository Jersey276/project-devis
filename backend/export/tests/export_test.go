package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"project-devis-export/actions/codes"
	quoteexport "project-devis-export/actions/quote"
	"project-devis-export/quote"
	exportGrpc "project-devis-export/services/grpc"
	"project-devis-export/users"
)

func validReq() *exportGrpc.ExportQuoteRequest {
	return &exportGrpc.ExportQuoteRequest{QuoteId: "quote-1", UserId: "user-1"}
}

func TestExport_InvalidInput(t *testing.T) {
	qc, uc, gt := happyFakes()

	for _, tc := range []struct {
		name string
		req  *exportGrpc.ExportQuoteRequest
	}{
		{"empty quote id", &exportGrpc.ExportQuoteRequest{UserId: "user-1"}},
		{"empty user id", &exportGrpc.ExportQuoteRequest{QuoteId: "quote-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := quoteexport.Export(context.Background(), qc, uc, gt, tc.req)
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if resp.Success || resp.Code != codes.InvalidInput {
				t.Fatalf("expected InvalidInput, got success=%v code=%d", resp.Success, resp.Code)
			}
		})
	}
}

func TestExport_QuoteNotFound(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(context.Context, *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{Success: false, Code: 1001}, nil // upstream NotFound
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Success || resp.Code != codes.NotFound {
		t.Fatalf("expected NotFound (%d), got success=%v code=%d", codes.NotFound, resp.Success, resp.Code)
	}
}

func TestExport_ClientNotFound(t *testing.T) {
	qc, uc, gt := happyFakes()
	uc.getClient = func(context.Context, *users.GetClientRequest) (*users.GetClientResponse, error) {
		return &users.GetClientResponse{Success: false, Code: 1001}, nil
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Success || resp.Code != codes.DependencyMissing {
		t.Fatalf("expected DependencyMissing (%d), got success=%v code=%d", codes.DependencyMissing, resp.Success, resp.Code)
	}
}

func TestExport_AddressNotFound(t *testing.T) {
	qc, uc, gt := happyFakes()
	uc.getAddress = func(context.Context, *users.GetAddressRequest) (*users.GetAddressResponse, error) {
		return &users.GetAddressResponse{Success: false, Code: 1001}, nil
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Success || resp.Code != codes.DependencyMissing {
		t.Fatalf("expected DependencyMissing (%d), got success=%v code=%d", codes.DependencyMissing, resp.Success, resp.Code)
	}
}

// A transport error from any upstream RPC must surface as (nil, err) so the
// gRPC framework turns it into a 502 at the gateway — not a structured
// success-shaped payload.
func TestExport_TransportError(t *testing.T) {
	qc, uc, gt := happyFakes()
	boom := errors.New("connection refused")
	qc.getQuote = func(context.Context, *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return nil, boom
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err == nil {
		t.Fatalf("expected transport error to propagate, got resp=%+v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response on transport error, got %+v", resp)
	}
	if !errors.Is(err, boom) {
		t.Fatalf("expected error to wrap %v, got %v", boom, err)
	}
}

func TestExport_RenderError(t *testing.T) {
	qc, uc, gt := happyFakes()
	gt.convert = func(context.Context, []byte) ([]byte, error) {
		return nil, errors.New("chromium crashed")
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err == nil {
		t.Fatalf("expected render error to propagate, got resp=%+v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response on render error, got %+v", resp)
	}
}

func TestExport_Success(t *testing.T) {
	qc, uc, gt := happyFakes()

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("expected Success, got success=%v code=%d", resp.Success, resp.Code)
	}
	if len(resp.Pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !strings.HasPrefix(resp.Filename, "devis-") || !strings.HasSuffix(resp.Filename, ".pdf") {
		t.Fatalf("expected filename like devis-*.pdf, got %q", resp.Filename)
	}
	// The fake quote name "Cuisine équipée" should slug to "cuisine-equipee" if
	// the slugifier diacritic-strips; if not, at least non-empty and ending .pdf.
	if resp.Filename == "devis-.pdf" {
		t.Fatalf("expected slug from quote name, got bare prefix: %q", resp.Filename)
	}
}
