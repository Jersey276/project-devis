package actions

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// genesisHash is the predecessor hash of the first document in an issuer's chain
// (64 lowercase zeros).
const genesisHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Canonical serialization separators: control bytes that cannot appear in user
// text, so a crafted line name can never forge an ambiguous serialization.
const (
	fieldSep  = "\x1f" // US — between fields
	recordSep = "\x1e" // RS — between line records
)

// sealLine is the frozen content of one document line that enters the hash.
type sealLine struct {
	name           string
	quantity       string
	unit           string
	unitPriceCents int64
	lineHTCents    int64
	taxRate        string // canonical string, never a parsed float
	taxLabel       string
}

// sealableDoc is the normalized legal content of an issued document, fed
// identically by the live seal path, the backfill, and the verifier.
type sealableDoc struct {
	userID              string
	docType             string // "INVOICE" | "CREDIT_NOTE"
	number              string // invoice_number or credit_note_number
	issuedAt            time.Time
	totalHT             int64
	totalVAT            int64
	totalTTC            int64
	vatExempt           bool
	originInvoiceNumber string // "" for invoices; origin invoice number for credit notes
	lines               []sealLine
}

// computeContentHash hashes the legal content of a document deterministically.
// This is a FROZEN wire format: changing it invalidates every existing seal and
// must be accompanied by a re-seal migration.
func computeContentHash(doc sealableDoc) string {
	var b strings.Builder

	// Header fields, fixed order.
	header := []string{
		doc.docType,
		doc.userID,
		doc.number,
		doc.issuedAt.UTC().Format(time.RFC3339Nano),
		strconv.FormatInt(doc.totalHT, 10),
		strconv.FormatInt(doc.totalVAT, 10),
		strconv.FormatInt(doc.totalTTC, 10),
		bool01(doc.vatExempt),
		doc.originInvoiceNumber,
	}
	b.WriteString(strings.Join(header, fieldSep))

	// Lines digest, in the order given (callers pass position order).
	b.WriteString(fieldSep)
	for i, l := range doc.lines {
		if i > 0 {
			b.WriteString(recordSep)
		}
		rec := []string{
			l.name,
			l.quantity,
			l.unit,
			strconv.FormatInt(l.unitPriceCents, 10),
			strconv.FormatInt(l.lineHTCents, 10),
			l.taxRate,
			l.taxLabel,
		}
		b.WriteString(strings.Join(rec, fieldSep))
	}

	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// computeChainHash links a document to its predecessor:
// sha256(prevHash || content || index).
func computeChainHash(prevHash, contentHash string, index int64) string {
	payload := prevHash + fieldSep + contentHash + fieldSep + strconv.FormatInt(index, 10)
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:])
}

func bool01(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

// sealDocument allocates the next chain link for the issuer and writes the seal,
// inside the caller's transaction. It locks the issuer's chain head FOR UPDATE so
// concurrent emissions of the same issuer serialise (no gap, no collision).
func sealDocument(ctx context.Context, tx *sql.Tx, userID, docType, docID, contentHash string) (chainIndex int64, chainHash string, err error) {
	var lastIndex int64
	var lastHash string
	err = tx.QueryRowContext(ctx,
		`SELECT last_index, last_hash FROM chain_heads WHERE user_id=$1 FOR UPDATE`,
		userID,
	).Scan(&lastIndex, &lastHash)

	var prevHash string
	switch err {
	case nil:
		chainIndex = lastIndex + 1
		prevHash = lastHash
	case sql.ErrNoRows:
		chainIndex = 0
		prevHash = genesisHash
	default:
		return 0, "", fmt.Errorf("lock chain head: %w", err)
	}

	chainHash = computeChainHash(prevHash, contentHash, chainIndex)

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO document_seals (user_id, doc_type, doc_id, chain_index, content_hash, prev_hash, chain_hash)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		userID, docType, docID, chainIndex, contentHash, prevHash, chainHash,
	); err != nil {
		return 0, "", fmt.Errorf("insert seal: %w", err)
	}

	if _, err = tx.ExecContext(ctx,
		`INSERT INTO chain_heads (user_id, last_index, last_hash) VALUES ($1, $2, $3)
		 ON CONFLICT (user_id) DO UPDATE SET last_index=$2, last_hash=$3`,
		userID, chainIndex, chainHash,
	); err != nil {
		return 0, "", fmt.Errorf("update chain head: %w", err)
	}

	return chainIndex, chainHash, nil
}

// sealLinesFromSnapshots maps frozen invoice/credit-note line snapshots to the
// canonical sealLine form (used by the live seal path and the verifier).
func sealLinesFromSnapshots(lines []lineSnapshot) []sealLine {
	out := make([]sealLine, len(lines))
	for i, l := range lines {
		out[i] = sealLine{
			name:           l.name,
			quantity:       l.quantity,
			unit:           l.unit,
			unitPriceCents: l.unitPriceCents,
			lineHTCents:    l.lineHTCents,
			taxRate:        l.taxRate,
			taxLabel:       l.taxLabel,
		}
	}
	return out
}

func sealLinesFromCreditNoteLines(lines []creditNoteLine) []sealLine {
	out := make([]sealLine, len(lines))
	for i, l := range lines {
		out[i] = sealLine{
			name:           l.name,
			quantity:       l.quantity,
			unit:           l.unit,
			unitPriceCents: l.unitPriceCents,
			lineHTCents:    l.lineHTCents,
			taxRate:        l.taxRate,
			taxLabel:       l.taxLabel,
		}
	}
	return out
}
