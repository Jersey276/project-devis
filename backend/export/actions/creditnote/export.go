package creditnote

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"project-devis-export/actions/codes"
	exportGrpc "project-devis-export/services/grpc"
	invoicepb "project-devis-export/services/invoice"
)

// Upstream invoice-service codes mirrored here (kept in sync with
// backend/invoice/actions/codes/codes.go) so we don't import across services.
const (
	upstreamNotFound     int32 = 1001
	upstreamInvalidInput int32 = 1003
)

type pdfConverter interface {
	Convert(ctx context.Context, html []byte) ([]byte, error)
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

	pdfBytes, err := Render(ctx, gt, resp.GetCreditNote())
	if err != nil {
		return nil, err
	}

	return &exportGrpc.ExportQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		Pdf:      pdfBytes,
		Filename: buildFilename(resp.GetCreditNote()),
	}, nil
}

func buildFilename(cn *invoicepb.CreditNoteDetails) string {
	if num := strings.TrimSpace(cn.GetCreditNoteNumber()); num != "" {
		return fmt.Sprintf("avoir-%s.pdf", slugify(num))
	}
	return fmt.Sprintf("avoir-%s.pdf", cn.GetCreditNoteId())
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
