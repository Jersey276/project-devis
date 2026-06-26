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

func TestExport_QuoteRefused(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{
			Success: true,
			Quote: &quote.Quote{
				QuoteId:   req.QuoteId,
				UserId:    req.UserId,
				Name:      "Devis refuse",
				State:     quote.QuoteState_QUOTE_STATE_DROP,
				ClientId:  "client-1",
				AddressId: 42,
			},
		}, nil
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Success || resp.Code != codes.QuoteRefused {
		t.Fatalf("expected QuoteRefused (%d), got success=%v code=%d", codes.QuoteRefused, resp.Success, resp.Code)
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

func TestExport_VATRendered(t *testing.T) {
	qc, uc, gt := happyFakes()
	var capturedHTML []byte
	gt.convert = func(_ context.Context, html []byte) ([]byte, error) {
		capturedHTML = html
		return []byte("%PDF-1.4 fake"), nil
	}

	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil || !resp.Success {
		t.Fatalf("export failed: err=%v resp=%+v", err, resp)
	}
	for _, want := range []string{"20.00 %", "Total TVA", "Total TTC"} {
		if !strings.Contains(string(capturedHTML), want) {
			t.Errorf("HTML output missing %q", want)
		}
	}
}

func captureHTML(t *testing.T, qc *fakeQuote, uc *fakeUsers, gt *fakeGotenberg) string {
	t.Helper()
	var capturedHTML []byte
	origConvert := gt.convert
	gt.convert = func(ctx context.Context, html []byte) ([]byte, error) {
		capturedHTML = html
		return origConvert(ctx, html)
	}
	resp, err := quoteexport.Export(context.Background(), qc, uc, gt, validReq())
	if err != nil || !resp.Success {
		t.Fatalf("export failed: err=%v resp=%+v", err, resp)
	}
	return string(capturedHTML)
}

func TestExport_GroupLine_RenderedAsHeader(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{
			Success: true,
			Quote:   &quote.Quote{QuoteId: req.QuoteId, UserId: req.UserId, Name: "Test", ClientId: "client-1", AddressId: 42},
			Lines: []*quote.QuoteLine{
				{LineId: "l1", Name: "Matériaux", Data: `{"kind":"group"}`},
				{LineId: "l2", Name: "Vis", Quantity: "10", UnitPrice: 100, TaxId: 1, Data: `{"kind":"line"}`},
			},
		}, nil
	}
	html := captureHTML(t, qc, uc, gt)
	if !strings.Contains(html, "group-header") {
		t.Error("HTML missing group-header class")
	}
	if !strings.Contains(html, "Matériaux") {
		t.Error("HTML missing group name")
	}
}

func TestExport_TextLine_RenderedFullWidth(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{
			Success: true,
			Quote:   &quote.Quote{QuoteId: req.QuoteId, UserId: req.UserId, Name: "Test", ClientId: "client-1", AddressId: 42},
			Lines: []*quote.QuoteLine{
				{LineId: "l1", Name: "Travaux réalisés selon les normes en vigueur.", Data: `{"kind":"text"}`},
				{LineId: "l2", Name: "Pose", Quantity: "1", UnitPrice: 50000, TaxId: 1},
			},
		}, nil
	}
	html := captureHTML(t, qc, uc, gt)
	if !strings.Contains(html, "text-line") {
		t.Error("HTML missing text-line class")
	}
	if !strings.Contains(html, "Travaux réalisés") {
		t.Error("HTML missing text content")
	}
}

func TestExport_OptionLine_ExcludedFromTotal(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{
			Success: true,
			Quote:   &quote.Quote{QuoteId: req.QuoteId, UserId: req.UserId, Name: "Test", ClientId: "client-1", AddressId: 42},
			Lines: []*quote.QuoteLine{
				// Normal line: 500,00 € HT
				{LineId: "l1", Name: "Pose", Quantity: "1", UnitPrice: 50000, TaxId: 1, Data: `{"kind":"line"}`},
				// Option line: 200,00 € HT (must NOT be in TotalHT)
				{LineId: "l2", Name: "Garantie étendue", Quantity: "1", UnitPrice: 20000, Data: `{"kind":"line","option":true}`},
			},
		}, nil
	}
	html := captureHTML(t, qc, uc, gt)
	// Options section must appear
	if !strings.Contains(html, "Options") {
		t.Error("HTML missing options section")
	}
	if !strings.Contains(html, "Garantie étendue") {
		t.Error("HTML missing option line name")
	}
	// TotalHT must reflect only the normal line (500,00 €), not 700,00 €
	if strings.Contains(html, "700,00") {
		t.Error("option line amount incorrectly included in total")
	}
}

func TestExport_DetailedLine_SublineExpanded(t *testing.T) {
	qc, uc, gt := happyFakes()
	qc.getQuote = func(_ context.Context, req *quote.GetQuoteRequest) (*quote.GetQuoteResponse, error) {
		return &quote.GetQuoteResponse{
			Success: true,
			Quote:   &quote.Quote{QuoteId: req.QuoteId, UserId: req.UserId, Name: "Test", ClientId: "client-1", AddressId: 42},
			Lines: []*quote.QuoteLine{
				{
					LineId: "l1", Name: "Installation complète", Type: "multiple", TaxId: 1,
					Data: `{"kind":"detailed","sublines":[{"name":"Fourniture matériel","quantity":"2","unit":"u","unit_price":15000},{"name":"Main d'oeuvre","quantity":"3","unit":"h","unit_price":5000}]}`,
				},
			},
		}, nil
	}
	html := captureHTML(t, qc, uc, gt)
	if !strings.Contains(html, "Fourniture matériel") {
		t.Error("HTML missing first subline")
	}
	if !strings.Contains(html, `class="subline"`) {
		t.Error("HTML missing subline class")
	}
	// Total: 2*150 + 3*50 = 300 + 150 = 450,00 €
	if !strings.Contains(html, "450,00") {
		t.Error("HTML missing correct detailed total")
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
