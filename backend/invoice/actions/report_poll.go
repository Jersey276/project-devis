package actions

import (
	"context"
	"log"
	"time"

	"github.com/lib/pq"

	"project-devis-invoice/pdp"
)

// PollReportStatuses reconciles each submitted, non-terminal e-report with the
// platform once (B5/C5 mirror of PollPDPStatuses). For every report carrying a
// report_id and sitting in a non-terminal main-path status, it asks the platform
// for the current status and, if further along, advances the report one cran at a
// time through the same strict B3 path (reconcileSteps). The default no-op reporter
// returns UNKNOWN, so the worker is inert in production until a real PA is wired.
// Errors on one report are logged and skipped; the sweep continues.
func (s *Server) PollReportStatuses(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, status, report_id
		   FROM invoice_reports
		  WHERE report_id IS NOT NULL
		    AND status = ANY($1)`,
		pq.Array(pollableLifecycleStatuses),
	)
	if err != nil {
		log.Printf("invoice report-poll: query failed: %v", err)
		return
	}
	defer rows.Close()

	type pending struct {
		id       int64
		status   string
		reportID string
	}
	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.id, &p.status, &p.reportID); err != nil {
			log.Printf("invoice report-poll: scan failed: %v", err)
			return
		}
		batch = append(batch, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("invoice report-poll: rows: %v", err)
		return
	}

	for _, p := range batch {
		platform, err := s.reporter.FetchReportStatus(ctx, p.reportID)
		if err != nil {
			log.Printf("invoice report-poll: fetch %d: %v", p.id, err)
			continue
		}
		target, ok := pdp.ToLifecycleStatus(platform)
		if !ok { // UNKNOWN / unmapped: leave the report as-is.
			continue
		}
		s.reconcileReport(ctx, p.id, p.status, target)
	}
}

// reconcileReport advances one report from current to target along the strict B3
// path, applying each planned cran in its own statement. Unlike invoices there is
// no append-only event log: the status lives on the report row, so a cran is a
// guarded UPDATE (WHERE status = previous) that is naturally idempotent.
func (s *Server) reconcileReport(ctx context.Context, id int64, current, target string) {
	prev := current
	for _, step := range reconcileSteps(current, target) {
		res, err := s.db.ExecContext(ctx,
			`UPDATE invoice_reports SET status=$1, updated_at=NOW()
			  WHERE id=$2 AND status=$3`,
			step, id, prev,
		)
		if err != nil {
			log.Printf("invoice report-poll: advance %d -> %s: %v", id, step, err)
			return // next sweep retries from the new state.
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return // concurrent advance: stop, the next sweep reconciles.
		}
		prev = step
	}
}

// StartReportPoller runs PollReportStatuses on a ticker until ctx is cancelled.
// interval of 0 disables the loop. Each sweep is bounded by a timeout so a stuck
// platform call cannot wedge the worker. Mirrors StartPDPPoller.
func (s *Server) StartReportPoller(ctx context.Context, interval, sweepTimeout time.Duration) {
	if interval <= 0 {
		return
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sweepCtx, cancel := context.WithTimeout(ctx, sweepTimeout)
			s.PollReportStatuses(sweepCtx)
			cancel()
		}
	}
}
