package schedule

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strings"
	"time"

	"project-devis-export/internal/format"
	schedulepb "project-devis-export/services/schedule"
	"project-devis-export/templates"
)

var scheduleTpl = template.Must(template.New("schedule.html").Parse(string(templates.ScheduleHTML)))

type scheduleRenderInput struct {
	Schedule  *schedulepb.ScheduleDetails
	QuoteName string
}

type scheduleViewModel struct {
	ShortID        string
	ScheduleName   string
	Status         string
	StartMonth     string
	DurationMonths int32
	QuoteLabel     string
	MonthlyTotals  []scheduleMonthTotalView
	PlannedTotal   string
}

type scheduleMonthTotalView struct {
	Label     string
	Amount    string
	Cumulated string
}

var scheduleStatusLabels = map[string]string{
	"DRAFT":     "Brouillon",
	"NEGOCIATE": "En négociation",
	"DENIED":    "Refusé",
	"VALID":     "Validé",
}

func scheduleStatusLabel(status string) string {
	if label, ok := scheduleStatusLabels[status]; ok {
		return label
	}
	return status
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

	amountByMonth := make(map[int32]int64, len(s.ColumnTotals))
	for _, col := range s.ColumnTotals {
		amountByMonth[col.MonthIndex] = col.AmountCents
	}

	months := make([]scheduleMonthTotalView, 0, s.DurationMonths)
	var cumulated int64
	for monthIndex := int32(1); monthIndex <= s.DurationMonths; monthIndex++ {
		cumulated += amountByMonth[monthIndex]
		months = append(months, scheduleMonthTotalView{
			Label:     monthLabel(s.StartMonth, monthIndex),
			Amount:    format.Cents(amountByMonth[monthIndex]),
			Cumulated: format.Cents(cumulated),
		})
	}

	quoteLabel := strings.TrimSpace(in.QuoteName)
	if quoteLabel == "" {
		quoteLabel = format.ShortID(s.QuoteId)
	}

	return scheduleViewModel{
		ShortID:        format.ShortID(s.ScheduleId),
		ScheduleName:   s.Name,
		Status:         scheduleStatusLabel(s.Status),
		StartMonth:     s.StartMonth,
		DurationMonths: s.DurationMonths,
		QuoteLabel:     quoteLabel,
		MonthlyTotals:  months,
		PlannedTotal:   format.Cents(s.PlannedTotalCents),
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
