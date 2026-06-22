package invoice

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"project-devis-export/actions/codes"
	"project-devis-export/actions/invoice/facturx"
	"project-devis-export/services/facturxpdf"
	invoicepb "project-devis-export/services/invoice"
	exportGrpc "project-devis-export/services/grpc"
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

func Export(ctx context.Context, ic invoicepb.InvoiceServiceClient, gt pdfConverter, req *exportGrpc.ExportInvoiceRequest) (*exportGrpc.ExportQuoteResponse, error) {
	if req.InvoiceId == "" || req.UserId == "" {
		return fail(codes.InvalidInput), nil
	}

	resp, err := ic.GetInvoice(ctx, &invoicepb.GetInvoiceRequest{InvoiceId: req.InvoiceId, UserId: req.UserId})
	if err != nil {
		return nil, err
	}
	if !resp.GetSuccess() || resp.GetInvoice() == nil {
		return fail(mapInvoiceCode(resp.GetCode())), nil
	}
	in := resp.GetInvoice()

	var pdfBytes []byte
	if req.GetFacturx() {
		// A Factur-X invoice requires a legal number: a draft has none, so refuse
		// early rather than render an invalid structured document.
		if in.GetStatus() == "DRAFT" || strings.TrimSpace(in.GetInvoiceNumber()) == "" {
			return fail(codes.InvalidInput), nil
		}
		pdfBytes, err = renderFacturx(ctx, gt, in)
	} else {
		pdfBytes, err = Render(ctx, gt, in)
	}
	if err != nil {
		return nil, err
	}

	return &exportGrpc.ExportQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		Pdf:      pdfBytes,
		Filename: buildFilename(in),
	}, nil
}

// renderFacturx builds the hybrid Factur-X PDF: a PDF/A-3 of the visual invoice
// with the EN 16931 CII XML embedded as factur-x.xml.
func renderFacturx(ctx context.Context, gt pdfConverter, in *invoicepb.InvoiceDetails) ([]byte, error) {
	xmlBytes, err := facturx.Build(in)
	if err != nil {
		return nil, err
	}
	pdfA3, err := RenderPDFA3(ctx, gt, in)
	if err != nil {
		return nil, err
	}
	return facturxpdf.Assemble(pdfA3, xmlBytes)
}

func buildFilename(in *invoicepb.InvoiceDetails) string {
	if num := strings.TrimSpace(in.GetInvoiceNumber()); num != "" {
		return fmt.Sprintf("facture-%s.pdf", slugify(num))
	}
	return fmt.Sprintf("facture-%s.pdf", in.GetInvoiceId())
}

func slugify(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		case unicode.IsSpace(r) || r == '_' || r == '-':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := b.String()
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return out
}

func fail(code int32) *exportGrpc.ExportQuoteResponse {
	return &exportGrpc.ExportQuoteResponse{Success: false, Code: code}
}

func mapInvoiceCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.NotFound
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}
