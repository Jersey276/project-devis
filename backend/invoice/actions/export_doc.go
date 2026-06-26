package actions

import (
	"context"
	"fmt"

	exportGrpc "project-devis-invoice/services/exportgrpc"
)

// ExportDocumentSource implements iopole.DocumentSource by calling the export
// service's ExportInvoice with the Factur-X flag, reusing the validated EN16931
// PDF/A-3 pipeline rather than re-generating the document in the PA adapter.
//
// Note: export is itself a client of invoice (export → invoice.GetInvoice), so this
// is a runtime back-call invoice → export. It is safe because the calls are distinct
// and non-reentrant (ExportInvoice does not call back into deposit).
type ExportDocumentSource struct {
	client exportGrpc.ExportServiceClient
}

func NewExportDocumentSource(client exportGrpc.ExportServiceClient) *ExportDocumentSource {
	return &ExportDocumentSource{client: client}
}

// FetchFacturX returns the Factur-X PDF/A-3 bytes for an issued invoice.
func (s *ExportDocumentSource) FetchFacturX(ctx context.Context, invoiceID, userID string) ([]byte, error) {
	resp, err := s.client.ExportInvoice(ctx, &exportGrpc.ExportInvoiceRequest{
		InvoiceId: invoiceID,
		UserId:    userID,
		Facturx:   true,
	})
	if err != nil {
		return nil, fmt.Errorf("export invoice: %w", err)
	}
	if !resp.GetSuccess() || len(resp.GetPdf()) == 0 {
		return nil, fmt.Errorf("export invoice failed: code=%d", resp.GetCode())
	}
	return resp.GetPdf(), nil
}
