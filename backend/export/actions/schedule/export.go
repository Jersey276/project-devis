package schedule

import (
	"context"
	"fmt"
	"strings"

	"project-devis-export/actions/codes"
	"project-devis-export/internal/slug"
	"project-devis-export/quote"
	exportGrpc "project-devis-export/services/grpc"
	schedulepb "project-devis-export/services/schedule"
)

const (
	upstreamNotFound     int32 = 1001
	upstreamInvalidInput int32 = 1003
)

type schedulePDFConverter interface {
	Convert(ctx context.Context, html []byte) ([]byte, error)
}

func Export(ctx context.Context, sc schedulepb.ScheduleServiceClient, qc quote.QuoteServiceClient,
	gt schedulePDFConverter, req *exportGrpc.ExportScheduleRequest) (*exportGrpc.ExportQuoteResponse, error) {

	if strings.TrimSpace(req.ScheduleId) == "" || strings.TrimSpace(req.UserId) == "" {
		return fail(codes.InvalidInput), nil
	}

	sResp, err := sc.GetSchedule(ctx, &schedulepb.GetScheduleRequest{
		ScheduleId: req.ScheduleId,
		UserId:     req.UserId,
	})
	if err != nil {
		return nil, err
	}
	if !sResp.Success || sResp.Schedule == nil {
		return fail(mapScheduleCode(sResp.Code)), nil
	}

	lineByID := map[string]*quote.QuoteLine{}
	qResp, err := qc.GetQuote(ctx, &quote.GetQuoteRequest{
		QuoteId: sResp.Schedule.QuoteId,
		UserId:  req.UserId,
	})
	if err != nil {
		return nil, err
	}
	if qResp.Success {
		for _, line := range qResp.Lines {
			lineByID[line.LineId] = line
		}
	}

	pdfBytes, err := renderSchedule(ctx, gt, scheduleRenderInput{
		Schedule:  sResp.Schedule,
		QuoteLine: lineByID,
	})
	if err != nil {
		return nil, err
	}

	return &exportGrpc.ExportQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		Pdf:      pdfBytes,
		Filename: buildFilename(sResp.Schedule),
	}, nil
}

func mapScheduleCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.NotFound
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}

func buildFilename(s *schedulepb.ScheduleDetails) string {
	name := slug.Slugify(s.Name)
	if name == "" {
		return fmt.Sprintf("echeancier-%s.pdf", s.ScheduleId)
	}
	return fmt.Sprintf("echeancier-%s.pdf", name)
}

func fail(code int32) *exportGrpc.ExportQuoteResponse {
	return &exportGrpc.ExportQuoteResponse{Success: false, Code: code}
}
