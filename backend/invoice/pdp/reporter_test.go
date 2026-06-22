package pdp

import (
	"context"
	"errors"
	"testing"
)

func TestNoopReporter_SubmitAcceptsLocally(t *testing.T) {
	res, err := NoopReporter{}.SubmitReport(context.Background(), SubmitReportInput{
		UserID: "u", Kind: ReportTransaction, Period: ReportPeriod{Year: 2099, Month: 4},
	})
	if err != nil {
		t.Fatalf("SubmitReport: %v", err)
	}
	if res.Status != PlatformSubmitted {
		t.Errorf("status=%q; want SUBMITTED", res.Status)
	}
	if res.ReportID != "" {
		t.Errorf("ReportID=%q; want empty (no-op assigns none)", res.ReportID)
	}
}

func TestNoopReporter_FetchStatusUnknown(t *testing.T) {
	st, err := NoopReporter{}.FetchReportStatus(context.Background(), "any")
	if err != nil {
		t.Fatalf("FetchReportStatus: %v", err)
	}
	if st != PlatformUnknown {
		t.Errorf("status=%q; want UNKNOWN (no-op never advances)", st)
	}
}

func TestMockReporter_RecordsAndReturnsCanned(t *testing.T) {
	m := &MockReporter{SubmitResult: SubmitReportResult{ReportID: "rep-1", Status: PlatformSubmitted}}
	in := SubmitReportInput{UserID: "u", Kind: ReportCrossBorderB2C, Period: ReportPeriod{Year: 2099, Month: 5}}
	res, err := m.SubmitReport(context.Background(), in)
	if err != nil {
		t.Fatalf("SubmitReport: %v", err)
	}
	if res.ReportID != "rep-1" {
		t.Errorf("ReportID=%q; want rep-1", res.ReportID)
	}
	if len(m.Reports) != 1 || m.Reports[0].Kind != ReportCrossBorderB2C {
		t.Errorf("recorded reports=%+v; want one CROSS_BORDER_B2C", m.Reports)
	}
}

func TestMockReporter_SubmitError(t *testing.T) {
	m := &MockReporter{SubmitErr: errors.New("platform down")}
	if _, err := m.SubmitReport(context.Background(), SubmitReportInput{}); err == nil {
		t.Fatal("SubmitReport: want error, got nil")
	}
}
