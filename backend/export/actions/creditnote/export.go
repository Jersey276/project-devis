package creditnote

import (
	"context"
	"fmt"
	"strings"

	"project-devis-export/actions/codes"
	"project-devis-export/actions/invoice/facturx"
	"project-devis-export/internal/slug"
	"project-devis-export/services/facturxpdf"
	exportGrpc "project-devis-export/services/grpc"
	invoicepb "project-devis-export/services/invoice"
)

// Upstream invoice-service codes mirrored here (kept in sync with
// backend/invoice/actions/codes/codes.go) so we don't import across services.
const (
	upstreamNotFound     int32 = 1001
	upstreamInvalidInput int32 = 1003
)

// pdfConverter renders the plain visual PDF and, for Factur-X, a PDF/A-3 base.
type pdfConverter interface {
	Convert(ctx context.Context, html []byte) ([]byte, error)
	ConvertPDFA3(ctx context.Context, html []byte) ([]byte, error)
}

func Export(ctx context.Context, ic invoicepb.InvoiceServiceClient, gt pdfConverter, req *exportGrpc.ExportCreditNoteRequest) (*exportGrpc.ExportQuoteResponse, error) {
	if req.CreditNoteId == "" || req.UserId == "" {
		return fail(codes.InvalidInput), nil
	}

	resp, err := ic.GetCreditNote(ctx, &invoicepb.GetCreditNoteRequest{CreditNoteId: req.CreditNoteId, UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	if !resp.GetSuccess() || resp.GetCreditNote() == nil {
		return fail(mapCode(resp.GetCode())), nil
	}
	cn := resp.GetCreditNote()

	var pdfBytes []byte
	if req.GetFacturx() {
		// A Factur-X credit note requires a legal number and a referenced invoice
		// number (BT-3); refuse early rather than render an invalid document.
		if strings.TrimSpace(cn.GetCreditNoteNumber()) == "" || strings.TrimSpace(cn.GetInvoiceNumber()) == "" {
			return fail(codes.InvalidInput), nil
		}
		pdfBytes, err = renderFacturx(ctx, gt, cn)
	} else {
		pdfBytes, err = Render(ctx, gt, cn)
	}
	if err != nil {
		return nil, err
	}

	return &exportGrpc.ExportQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		Pdf:      pdfBytes,
		Filename: buildFilename(cn),
	}, nil
}

// renderFacturx builds the hybrid Factur-X PDF: a PDF/A-3 of the visual credit
// note with the EN 16931 CII XML (type 381) embedded as factur-x.xml.
func renderFacturx(ctx context.Context, gt pdfConverter, cn *invoicepb.CreditNoteDetails) ([]byte, error) {
	xmlBytes, err := facturx.BuildCreditNote(cn)
	if err != nil {
		return nil, err
	}
	pdfA3, err := RenderPDFA3(ctx, gt, cn)
	if err != nil {
		return nil, err
	}
	return facturxpdf.Assemble(pdfA3, xmlBytes)
}

func buildFilename(cn *invoicepb.CreditNoteDetails) string {
	if num := strings.TrimSpace(cn.GetCreditNoteNumber()); num != "" {
		return fmt.Sprintf("avoir-%s.pdf", slug.Slugify(num))
	}
	return fmt.Sprintf("avoir-%s.pdf", cn.GetCreditNoteId())
}

func fail(code int32) *exportGrpc.ExportQuoteResponse {
	return &exportGrpc.ExportQuoteResponse{Success: false, Code: code}
}

func mapCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.NotFound
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}
