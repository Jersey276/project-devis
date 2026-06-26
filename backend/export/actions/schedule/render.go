package schedule

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"project-devis-export/internal/format"
	"project-devis-export/quote"
	schedulepb "project-devis-export/services/schedule"
	"project-devis-export/templates"
)

var scheduleTpl = template.Must(template.New("schedule.html").Parse(string(templates.ScheduleHTML)))

type scheduleRenderInput struct {
	Schedule  *schedulepb.ScheduleDetails
	QuoteLine map[string]*quote.QuoteLine
}

type scheduleViewModel struct {
	ShortID        string
	ScheduleName   string
	Status         string
	StartMonth     string
	DurationMonths int32
	QuoteID        string
	Lines          []scheduleLineView
	MonthlyTotals  []scheduleMonthTotalView
	PlannedTotal   string
	QuoteTotal     string
}

type scheduleLineView struct {
	Name      string
	Expected  string
	Planned   string
	Remaining string
}

type scheduleMonthTotalView struct {
	Label  string
	Amount string
}

func renderSchedule(ctx context.Context, gt schedulePDFConverter, in scheduleRenderInput) ([]byte, error) {
	vm := buildScheduleViewModel(in)

	var html bytes.Buffer
	if err := scheduleTpl.Execute(&html, vm); err != nil {
		return nil, fmt.Errorf("render schedule template: %w", err)
	}
	return gt.Convert(ctx, html.Bytes())
}

func buildScheduleViewModel(in scheduleRenderInput) scheduleViewModel {
	s := in.Schedule
	lines := make([]scheduleLineView, 0, len(s.Lines))
	for _, line := range s.Lines {
		name := line.QuoteLineId
		if ql, ok := in.QuoteLine[line.QuoteLineId]; ok && strings.TrimSpace(ql.Name) != "" {
			name = ql.Name
		}
		lines = append(lines, scheduleLineView{
			Name:      name,
			Expected:  format.Cents(line.ExpectedCents),
			Planned:   format.Cents(line.PlannedCents),
			Remaining: format.Cents(line.ExpectedCents - line.PlannedCents),
		})
	}

	months := make([]scheduleMonthTotalView, 0, len(s.ColumnTotals))
	for _, col := range s.ColumnTotals {
		months = append(months, scheduleMonthTotalView{
			Label:  monthLabel(s.StartMonth, col.MonthIndex),
			Amount: format.Cents(col.AmountCents),
		})
	}

	return scheduleViewModel{
		ShortID:        format.ShortID(s.ScheduleId),
		ScheduleName:   s.Name,
		Status:         s.Status,
		StartMonth:     s.StartMonth,
		DurationMonths: s.DurationMonths,
		QuoteID:        s.QuoteId,
		Lines:          lines,
		MonthlyTotals:  months,
		PlannedTotal:   format.Cents(s.PlannedTotalCents),
		QuoteTotal:     format.Cents(s.QuoteTotalCents),
	}
}

func monthLabel(startMonth string, monthIndex int32) string {
	base, err := time.Parse("2006-01", startMonth)
	if err != nil || monthIndex <= 0 {
		return fmt.Sprintf("Mois %d", monthIndex)
	}
	current := base.AddDate(0, int(monthIndex)-1, 0)
	return strings.Title(strings.ToLower(current.Format("Jan 2006")))
}

