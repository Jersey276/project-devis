package actions

import (
	"database/sql"
	"log"
	"time"
)

// StartPurgeWorker deletes rows older than 6 months every 24 hours.
// Uses a separate DB connection with DELETE-only privileges (purgeDB).
// If purgeDB is nil, purge is skipped (non-fatal — audit still functions).
func StartPurgeWorker(purgeDB *sql.DB) {
	if purgeDB == nil {
		log.Println("audit purge: PURGE_DB_USER not configured, skipping automatic purge")
		return
	}
	go func() {
		for {
			n, err := runPurge(purgeDB)
			if err != nil {
				log.Printf("audit purge: error: %v", err)
			} else {
				log.Printf("audit purge: deleted %d rows older than 6 months", n)
			}
			time.Sleep(24 * time.Hour)
		}
	}()
}

func runPurge(db *sql.DB) (int64, error) {
	res, err := db.Exec(`DELETE FROM activity_logs WHERE created_at < now() - INTERVAL '6 months'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
