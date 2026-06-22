package actions

import (
	"context"
	"log"
	"time"

	"github.com/lib/pq"

	"project-devis-invoice/actions/codes"
	"project-devis-invoice/pdp"
)

// lifecycleMainPath is the strict forward path the poller walks (REJECTED/NONE
// excluded: not on the linear advance path). The slice order is the rank, so the
// poller can tell how many crans separate the current status from a platform-
// reported one — B3 forbids skips, so each cran is applied in turn.
var lifecycleMainPath = []string{"DEPOSITED", "RECEIVED", "APPROVED", "COLLECTED"}

// pollableLifecycleStatuses are the non-terminal main-path statuses whose
// invoices the poller reconciles against the platform (everything but the
// terminal COLLECTED).
var pollableLifecycleStatuses = lifecycleMainPath[:len(lifecycleMainPath)-1]

// mainPathRank returns the index of status on lifecycleMainPath, or -1 if it is
// off the path (REJECTED, NONE, unknown).
func mainPathRank(status string) int {
	for i, s := range lifecycleMainPath {
		if s == status {
			return i
		}
	}
	return -1
}

// PollPDPStatuses reconciles each deposited, non-terminal invoice's lifecycle
// with the platform once (B6). For every invoice carrying a pdp_submission_id
// and sitting in DEPOSITED|RECEIVED|APPROVED, it asks the PA for the current
// status and, if that status is further along the flow, advances the lifecycle
// one cran at a time through applyLifecycleTransition — so the B3 guards, the
// strict no-skip table and the append-only event log all apply. The default
// no-op adapter returns UNKNOWN, leaving everything untouched: the worker is
// inert in production until a real PA is wired. Errors on one invoice are logged
// and skipped; the sweep continues.
func (s *Server) PollPDPStatuses(ctx context.Context) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT invoice_id, user_id, lifecycle_status, pdp_submission_id
		   FROM invoices
		  WHERE pdp_submission_id IS NOT NULL
		    AND lifecycle_status = ANY($1)`,
		pq.Array(pollableLifecycleStatuses),
	)
	if err != nil {
		log.Printf("invoice pdp-poll: query failed: %v", err)
		return
	}
	defer rows.Close()

	type pending struct {
		invoiceID, userID, lifecycle, submissionID string
	}
	var batch []pending
	for rows.Next() {
		var p pending
		if err := rows.Scan(&p.invoiceID, &p.userID, &p.lifecycle, &p.submissionID); err != nil {
			log.Printf("invoice pdp-poll: scan failed: %v", err)
			return
		}
		batch = append(batch, p)
	}
	if err := rows.Err(); err != nil {
		log.Printf("invoice pdp-poll: rows: %v", err)
		return
	}

	for _, p := range batch {
		platform, err := s.pdpClient.FetchStatus(ctx, p.submissionID)
		if err != nil {
			log.Printf("invoice pdp-poll: fetch %s: %v", p.invoiceID, err)
			continue
		}
		target, ok := pdp.ToLifecycleStatus(platform)
		if !ok { // UNKNOWN / unmapped: leave the invoice as-is.
			continue
		}
		s.reconcileLifecycle(ctx, p.invoiceID, p.userID, p.lifecycle, target)
	}
}

// reconcileSteps plans the lifecycle moves to take an invoice from current to a
// platform-reported target along the strict B3 path. REJECTED is a single direct
// move (reachable from any active state). A forward main-path target yields one
// step per cran (no skips). A target equal to or behind current — or off-path —
// yields no steps (platform lag, already reconciled, or terminal current).
func reconcileSteps(current, target string) []string {
	if target == current {
		return nil
	}
	from := mainPathRank(current)
	if from < 0 {
		return nil // current is off-path (REJECTED/NONE): nothing the poller can do.
	}
	if target == "REJECTED" {
		// Reachable from any active state, but COLLECTED is terminal.
		if current == "COLLECTED" {
			return nil
		}
		return []string{"REJECTED"}
	}
	to := mainPathRank(target)
	if to < 0 || to <= from {
		return nil // off-path target, or platform lagging behind local state.
	}
	return lifecycleMainPath[from+1 : to+1]
}

// reconcileLifecycle advances one invoice from current to target along the strict
// B3 path, applying each planned cran in its own tx through the shared guards.
func (s *Server) reconcileLifecycle(ctx context.Context, invoiceID, userID, current, target string) {
	for _, step := range reconcileSteps(current, target) {
		if err := s.applyOneReconcileStep(ctx, invoiceID, userID, step); err != nil {
			log.Printf("invoice pdp-poll: advance %s -> %s: %v", invoiceID, step, err)
			return // stop on first failure; next sweep retries from the new state.
		}
	}
}

// applyOneReconcileStep runs a single lifecycle transition in its own tx, so each
// cran is committed independently and a later failure does not roll back earlier
// progress. The note marks the move as platform-driven for the audit trail.
func (s *Server) applyOneReconcileStep(ctx context.Context, invoiceID, userID, target string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	code, err := applyLifecycleTransition(ctx, tx, invoiceID, userID, target, "PDP status sync")
	if err != nil {
		return err
	}
	if code != codes.Success {
		// A non-success business code here means the local state already moved
		// (concurrent advance) or the invoice left ISSUED|PAID; not a hard error.
		return nil
	}
	return tx.Commit()
}

// StartPDPPoller runs PollPDPStatuses on a ticker until ctx is cancelled. interval
// of 0 disables the loop (caller decided not to poll). Each sweep is bounded by a
// timeout so a stuck platform call cannot wedge the worker.
func (s *Server) StartPDPPoller(ctx context.Context, interval, sweepTimeout time.Duration) {
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
			s.PollPDPStatuses(sweepCtx)
			cancel()
		}
	}
}
