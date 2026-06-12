package tests

import (
	"context"
	"errors"
	"strings"
	"testing"

	"project-devis-export/actions/codes"
	scheduleexport "project-devis-export/actions/schedule"
	exportGrpc "project-devis-export/services/grpc"
	schedulepb "project-devis-export/services/schedule"
)

func validScheduleReq() *exportGrpc.ExportScheduleRequest {
	return &exportGrpc.ExportScheduleRequest{ScheduleId: "sch-1", UserId: "user-1"}
}

func happyScheduleFakes() (*fakeSchedule, *fakeQuote, *fakeGotenberg) {
	sc := &fakeSchedule{
		getSchedule: func(_ context.Context, req *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
			return &schedulepb.GetScheduleResponse{
				Success: true,
				Code:    0,
				Schedule: &schedulepb.ScheduleDetails{
					ScheduleId:        req.ScheduleId,
					QuoteId:           "quote-1",
					Status:            "DRAFT",
					Name:              "Planning principal",
					StartMonth:        "2026-06",
					DurationMonths:    3,
					Lines:             []*schedulepb.ScheduleLineSummary{{QuoteLineId: "l1", PlannedCents: 50000, ExpectedCents: 90000}},
					ColumnTotals:      []*schedulepb.ScheduleColumnTotal{{MonthIndex: 1, AmountCents: 50000}},
					QuoteTotalCents:   90000,
					PlannedTotalCents: 50000,
				},
			}, nil
		},
	}
	qc, _, gt := happyFakes()
	return sc, qc, gt
}

func TestExportSchedule_InvalidInput(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	for _, tc := range []struct {
		name string
		req  *exportGrpc.ExportScheduleRequest
	}{
		{"empty schedule id", &exportGrpc.ExportScheduleRequest{UserId: "user-1"}},
		{"empty user id", &exportGrpc.ExportScheduleRequest{ScheduleId: "sch-1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, tc.req)
			if err != nil {
				t.Fatalf("unexpected transport error: %v", err)
			}
			if resp.Success || resp.Code != codes.InvalidInput {
				t.Fatalf("expected InvalidInput, got success=%v code=%d", resp.Success, resp.Code)
			}
		})
	}
}

func TestExportSchedule_NotFound(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	sc.getSchedule = func(context.Context, *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
		return &schedulepb.GetScheduleResponse{Success: false, Code: 1001}, nil
	}

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if resp.Success || resp.Code != codes.NotFound {
		t.Fatalf("expected NotFound (%d), got success=%v code=%d", codes.NotFound, resp.Success, resp.Code)
	}
}

func TestExportSchedule_TransportError(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	boom := errors.New("schedule unavailable")
	sc.getSchedule = func(context.Context, *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
		return nil, boom
	}

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err == nil {
		t.Fatalf("expected transport error to propagate, got resp=%+v", resp)
	}
	if resp != nil {
		t.Fatalf("expected nil response on transport error, got %+v", resp)
	}
}

func TestExportSchedule_Success(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("expected Success, got success=%v code=%d", resp.Success, resp.Code)
	}
	if len(resp.Pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
	if !strings.HasPrefix(resp.Filename, "echeancier-") || !strings.HasSuffix(resp.Filename, ".pdf") {
		t.Fatalf("expected filename like echeancier-*.pdf, got %q", resp.Filename)
	}
	if !strings.Contains(resp.Filename, "Planning-principal") {
		t.Fatalf("expected schedule name in filename, got %q", resp.Filename)
	}
}

func TestExportSchedule_Success_WithEmptyLines(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	sc.getSchedule = func(_ context.Context, req *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
		return &schedulepb.GetScheduleResponse{
			Success: true,
			Code:    0,
			Schedule: &schedulepb.ScheduleDetails{
				ScheduleId:        req.ScheduleId,
				QuoteId:           "quote-1",
				Status:            "DRAFT",
				Name:              "Sans lignes",
				StartMonth:        "2026-06",
				DurationMonths:    3,
				Lines:             []*schedulepb.ScheduleLineSummary{},
				ColumnTotals:      []*schedulepb.ScheduleColumnTotal{{MonthIndex: 1, AmountCents: 0}},
				QuoteTotalCents:   0,
				PlannedTotalCents: 0,
			},
		}, nil
	}

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("expected Success, got success=%v code=%d", resp.Success, resp.Code)
	}
	if len(resp.Pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestExportSchedule_Success_WithEmptyMonthlyTotals(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	sc.getSchedule = func(_ context.Context, req *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
		return &schedulepb.GetScheduleResponse{
			Success: true,
			Code:    0,
			Schedule: &schedulepb.ScheduleDetails{
				ScheduleId:        req.ScheduleId,
				QuoteId:           "quote-1",
				Status:            "DRAFT",
				Name:              "Sans colonnes",
				StartMonth:        "2026-06",
				DurationMonths:    3,
				Lines:             []*schedulepb.ScheduleLineSummary{{QuoteLineId: "l1", PlannedCents: 0, ExpectedCents: 10000}},
				ColumnTotals:      []*schedulepb.ScheduleColumnTotal{},
				QuoteTotalCents:   10000,
				PlannedTotalCents: 0,
			},
		}, nil
	}

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("expected Success, got success=%v code=%d", resp.Success, resp.Code)
	}
	if len(resp.Pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}

func TestExportSchedule_Success_WhenDenied(t *testing.T) {
	sc, qc, gt := happyScheduleFakes()
	sc.getSchedule = func(_ context.Context, req *schedulepb.GetScheduleRequest) (*schedulepb.GetScheduleResponse, error) {
		return &schedulepb.GetScheduleResponse{
			Success: true,
			Code:    0,
			Schedule: &schedulepb.ScheduleDetails{
				ScheduleId:        req.ScheduleId,
				QuoteId:           "quote-1",
				Status:            "DENIED",
				Name:              "Version refusee",
				StartMonth:        "2026-06",
				DurationMonths:    3,
				Lines:             []*schedulepb.ScheduleLineSummary{{QuoteLineId: "l1", PlannedCents: 10000, ExpectedCents: 10000}},
				ColumnTotals:      []*schedulepb.ScheduleColumnTotal{{MonthIndex: 1, AmountCents: 10000}},
				QuoteTotalCents:   10000,
				PlannedTotalCents: 10000,
			},
		}, nil
	}

	resp, err := scheduleexport.Export(context.Background(), sc, qc, gt, validScheduleReq())
	if err != nil {
		t.Fatalf("unexpected transport error: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("expected Success for DENIED export, got success=%v code=%d", resp.Success, resp.Code)
	}
	if len(resp.Pdf) == 0 {
		t.Fatal("expected non-empty PDF bytes")
	}
}
