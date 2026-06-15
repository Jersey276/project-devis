package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/grpc"

	invoiceexport "project-devis-export/actions/invoice"
	exportGrpc "project-devis-export/services/grpc"
	invoicepb "project-devis-export/services/invoice"
)

// loadPDFA3Fixture returns a real PDF/A-3b document so facturxpdf.Assemble has
// valid input in the Factur-X export test.
func loadPDFA3Fixture(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "services", "facturxpdf", "testdata", "sample_pdfa3.pdf"))
	if err != nil {
		t.Fatalf("read PDF/A-3 fixture: %v", err)
	}
	return b
}

// fakeInvoice implements just the GetInvoice call the export orchestrator uses.
type fakeInvoice struct {
	invoicepb.InvoiceServiceClient // embedded: other methods panic if ever called
	getInvoice func(context.Context, *invoicepb.GetInvoiceRequest) (*invoicepb.GetInvoiceResponse, error)
}

func (f *fakeInvoice) GetInvoice(ctx context.Context, in *invoicepb.GetInvoiceRequest, _ ...grpc.CallOption) (*invoicepb.GetInvoiceResponse, error) {
	return f.getInvoice(ctx, in)
}

func issuedInvoiceDetails() *invoicepb.InvoiceDetails {
	return &invoicepb.InvoiceDetails{
		InvoiceId:     "inv-1",
		Status:        "ISSUED",
		InvoiceNumber: "2026-0001",
		IssuedAt:      "2026-06-14T10:30:00Z",
		Issuer:        &invoicepb.InvoiceParty{Company: "Acme SARL", Siren: "123456782"},
		Client:        &invoicepb.InvoiceParty{Company: "Buyer SAS", Siren: "987654321"},
		Lines: []*invoicepb.InvoiceLine{
			{Name: "Presta", Quantity: "1", UnitPriceCents: 10000, LineHtCents: 10000, TaxRate: "20"},
		},
		VatBreakdown:  []*invoicepb.InvoiceVatLine{{TaxRate: "20", BaseHtCents: 10000, VatCents: 2000}},
		TotalHtCents:  10000,
		TotalVatCents: 2000,
		TotalTtcCents: 12000,
	}
}

func fakeInvoiceReturning(d *invoicepb.InvoiceDetails) *fakeInvoice {
	return &fakeInvoice{
		getInvoice: func(_ context.Context, req *invoicepb.GetInvoiceRequest) (*invoicepb.GetInvoiceResponse, error) {
			return &invoicepb.GetInvoiceResponse{Success: true, Invoice: d}, nil
		},
	}
}

func TestExportInvoice_PlainPDF(t *testing.T) {
	ic := fakeInvoiceReturning(issuedInvoiceDetails())
	gt := &fakeGotenberg{convert: func(context.Context, []byte) ([]byte, error) { return []byte("%PDF-1.4 plain"), nil }}

	resp, err := invoiceexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportInvoiceRequest{InvoiceId: "inv-1", UserId: "u1", Facturx: false})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if !resp.Success {
		t.Fatalf("plain export failed: code=%d", resp.Code)
	}
	if gt.pdfa3Calls != 0 {
		t.Errorf("plain export must not call ConvertPDFA3 (got %d)", gt.pdfa3Calls)
	}
	if gt.convertCalls != 1 {
		t.Errorf("plain export should call Convert once (got %d)", gt.convertCalls)
	}
}

func TestExportInvoice_Facturx(t *testing.T) {
	ic := fakeInvoiceReturning(issuedInvoiceDetails())
	// ConvertPDFA3 returns a real PDF/A-3 fixture so facturxpdf.Assemble succeeds.
	gt := &fakeGotenberg{
		convert:      func(context.Context, []byte) ([]byte, error) { return []byte("%PDF-1.4 plain"), nil },
		convertPDFA3: func(context.Context, []byte) ([]byte, error) { return loadPDFA3Fixture(t), nil },
	}

	resp, err := invoiceexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportInvoiceRequest{InvoiceId: "inv-1", UserId: "u1", Facturx: true})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if !resp.Success {
		t.Fatalf("facturx export failed: code=%d", resp.Code)
	}
	if gt.pdfa3Calls != 1 {
		t.Errorf("facturx export should call ConvertPDFA3 once (got %d)", gt.pdfa3Calls)
	}
	if !bytes.HasPrefix(resp.Pdf, []byte("%PDF")) {
		t.Errorf("facturx output is not a PDF")
	}
}

func TestExportInvoice_FacturxRejectsDraft(t *testing.T) {
	d := issuedInvoiceDetails()
	d.Status = "DRAFT"
	d.InvoiceNumber = ""
	ic := fakeInvoiceReturning(d)
	gt := &fakeGotenberg{convert: func(context.Context, []byte) ([]byte, error) { return []byte("%PDF"), nil }}

	resp, err := invoiceexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportInvoiceRequest{InvoiceId: "inv-1", UserId: "u1", Facturx: true})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if resp.Success {
		t.Fatal("facturx export of a draft must fail")
	}
	if gt.pdfa3Calls != 0 || gt.convertCalls != 0 {
		t.Errorf("draft facturx must not render anything (pdfa3=%d convert=%d)", gt.pdfa3Calls, gt.convertCalls)
	}
}
