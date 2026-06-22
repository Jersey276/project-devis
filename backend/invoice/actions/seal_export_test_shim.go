package actions

import (
	"context"
	"database/sql"
	"time"
)

type SealLineForTest struct {
	Name           string
	Quantity       string
	Unit           string
	UnitPriceCents int64
	LineHTCents    int64
	TaxRate        string
	TaxLabel       string
}

type SealableDocForTest struct {
	UserID              string
	DocType             string
	Number              string
	IssuedAt            time.Time
	TotalHT             int64
	TotalVAT            int64
	TotalTTC            int64
	VatExempt           bool
	OriginInvoiceNumber string
	Lines               []SealLineForTest
}

func ComputeContentHashForTest(d SealableDocForTest) string {
	lines := make([]sealLine, len(d.Lines))
	for i, l := range d.Lines {
		lines[i] = sealLine{
			name:           l.Name,
			quantity:       l.Quantity,
			unit:           l.Unit,
			unitPriceCents: l.UnitPriceCents,
			lineHTCents:    l.LineHTCents,
			taxRate:        l.TaxRate,
			taxLabel:       l.TaxLabel,
		}
	}
	return computeContentHash(sealableDoc{
		userID:              d.UserID,
		docType:             d.DocType,
		number:              d.Number,
		issuedAt:            d.IssuedAt,
		totalHT:             d.TotalHT,
		totalVAT:            d.TotalVAT,
		totalTTC:            d.TotalTTC,
		vatExempt:           d.VatExempt,
		originInvoiceNumber: d.OriginInvoiceNumber,
		lines:               lines,
	})
}

func ComputeChainHashForTest(prevHash, contentHash string, index int64) string {
	return computeChainHash(prevHash, contentHash, index)
}

func GenesisHashForTest() string { return genesisHash }

func SealDocumentForTest(ctx context.Context, tx *sql.Tx, userID, docType, docID, contentHash string) (int64, string, error) {
	return sealDocument(ctx, tx, userID, docType, docID, contentHash)
}
