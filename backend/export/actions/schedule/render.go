package schedule

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

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
			Expected:  formatCents(line.ExpectedCents),
			Planned:   formatCents(line.PlannedCents),
			Remaining: formatCents(line.ExpectedCents - line.PlannedCents),
		})
	}

	months := make([]scheduleMonthTotalView, 0, len(s.ColumnTotals))
	for _, col := range s.ColumnTotals {
		months = append(months, scheduleMonthTotalView{
			Label:  monthLabel(s.StartMonth, col.MonthIndex),
			Amount: formatCents(col.AmountCents),
		})
	}

	return scheduleViewModel{
		ShortID:        shortID(s.ScheduleId),
		ScheduleName:   s.Name,
		Status:         s.Status,
		StartMonth:     s.StartMonth,
		DurationMonths: s.DurationMonths,
		QuoteID:        s.QuoteId,
		Lines:          lines,
		MonthlyTotals:  months,
		PlannedTotal:   formatCents(s.PlannedTotalCents),
		QuoteTotal:     formatCents(s.QuoteTotalCents),
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

func formatCents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	euros := cents / 100
	rem := cents % 100
	euroStr := groupThousands(strconv.FormatInt(euros, 10))
	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%s,%02d €", sign, euroStr, rem)
}

func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if n > pre {
			b.WriteByte(' ')
		}
	}
	for i := pre; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

func shortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
