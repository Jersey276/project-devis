package tests

import (
	"bytes"
	"context"
	"testing"

	"google.golang.org/grpc"

	creditnoteexport "project-devis-export/actions/creditnote"
	exportGrpc "project-devis-export/services/grpc"
	invoicepb "project-devis-export/services/invoice"
)

// fakeCreditNote implements just the GetCreditNote call the export orchestrator
// uses; other methods panic if ever called.
type fakeCreditNote struct {
	invoicepb.InvoiceServiceClient
	getCreditNote func(context.Context, *invoicepb.GetCreditNoteRequest) (*invoicepb.GetCreditNoteResponse, error)
}

func (f *fakeCreditNote) GetCreditNote(ctx context.Context, in *invoicepb.GetCreditNoteRequest, _ ...grpc.CallOption) (*invoicepb.GetCreditNoteResponse, error) {
	return f.getCreditNote(ctx, in)
}

func issuedCreditNoteDetails() *invoicepb.CreditNoteDetails {
	return &invoicepb.CreditNoteDetails{
		CreditNoteId:     "cn-1",
		InvoiceId:        "inv-1",
		InvoiceNumber:    "2026-0001",
		CreditNoteNumber: "AV-2026-0001",
		IssuedAt:         "2026-06-15T09:00:00Z",
		Reason:           "Annulation partielle",
		Issuer:           &invoicepb.InvoiceParty{Company: "Acme SARL", Siren: "123456782"},
		Client:           &invoicepb.InvoiceParty{Company: "Buyer SAS", Siren: "987654321"},
		Lines: []*invoicepb.InvoiceLine{
			{Name: "Presta", Quantity: "1", UnitPriceCents: 10000, LineHtCents: 10000, TaxRate: "20"},
		},
		VatBreakdown:  []*invoicepb.InvoiceVatLine{{TaxRate: "20", BaseHtCents: 10000, VatCents: 2000}},
		TotalHtCents:  10000,
		TotalVatCents: 2000,
		TotalTtcCents: 12000,
	}
}

func fakeCreditNoteReturning(d *invoicepb.CreditNoteDetails) *fakeCreditNote {
	return &fakeCreditNote{
		getCreditNote: func(_ context.Context, req *invoicepb.GetCreditNoteRequest) (*invoicepb.GetCreditNoteResponse, error) {
			return &invoicepb.GetCreditNoteResponse{Success: true, CreditNote: d}, nil
		},
	}
}

func TestExportCreditNote_PlainPDF(t *testing.T) {
	ic := fakeCreditNoteReturning(issuedCreditNoteDetails())
	gt := &fakeGotenberg{convert: func(context.Context, []byte) ([]byte, error) { return []byte("%PDF-1.4 plain"), nil }}

	resp, err := creditnoteexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportCreditNoteRequest{CreditNoteId: "cn-1", UserId: "u1", Facturx: false})
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

func TestExportCreditNote_Facturx(t *testing.T) {
	ic := fakeCreditNoteReturning(issuedCreditNoteDetails())
	gt := &fakeGotenberg{
		convert:      func(context.Context, []byte) ([]byte, error) { return []byte("%PDF-1.4 plain"), nil },
		convertPDFA3: func(context.Context, []byte) ([]byte, error) { return loadPDFA3Fixture(t), nil },
	}

	resp, err := creditnoteexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportCreditNoteRequest{CreditNoteId: "cn-1", UserId: "u1", Facturx: true})
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

// A credit note with no referenced invoice number cannot carry BT-3 and must be
// refused before any rendering happens.
func TestExportCreditNote_FacturxRejectsMissingInvoiceRef(t *testing.T) {
	d := issuedCreditNoteDetails()
	d.InvoiceNumber = ""
	ic := fakeCreditNoteReturning(d)
	gt := &fakeGotenberg{convert: func(context.Context, []byte) ([]byte, error) { return []byte("%PDF"), nil }}

	resp, err := creditnoteexport.Export(context.Background(), ic, gt,
		&exportGrpc.ExportCreditNoteRequest{CreditNoteId: "cn-1", UserId: "u1", Facturx: true})
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if resp.Success {
		t.Fatal("facturx export without a referenced invoice number must fail")
	}
	if gt.pdfa3Calls != 0 || gt.convertCalls != 0 {
		t.Errorf("rejected facturx must not render anything (pdfa3=%d convert=%d)", gt.pdfa3Calls, gt.convertCalls)
	}
}
