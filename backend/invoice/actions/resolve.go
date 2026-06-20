package actions

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sync/errgroup"

	"project-devis-invoice/actions/codes"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

// partySnapshot is the frozen legal-mentions block copied at issue time.
type partySnapshot struct {
	issuerCompany    string
	issuerSiren      string
	issuerVat        string
	issuerEmail      string
	issuerPhone      string
	issuerLogoURL    string
	issuerStreet     string
	issuerAdditional string
	issuerZip        string
	issuerCity       string
	clientFirstName  string
	clientLastName   string
	clientCompany    string
	clientSiren      string
	clientVat        string
	clientEmail      string
	clientStreet     string
	clientAdditional string
	clientZip        string
	clientCity       string
	clientType       string // frozen B2C/B2B nature: "individual" / "business"
	clientCountryID  int32  // frozen client country id (0 = unknown); drives OSS
	ossApplied       bool   // frozen: destination-country VAT (OSS) was applied at issue
}

// lineSnapshot is one frozen invoice line (mirrors the DB row / proto message).
type lineSnapshot struct {
	position       int32
	quoteLineID    string
	name           string
	unit           string
	quantity       string
	unitPriceCents int64
	lineHTCents    int64
	taxID          int32
	taxRate        string
	taxLabel       string
}

// resolvedInvoice carries everything needed to compute totals and write the
// snapshot, fully resolved from the downstream services.
type resolvedInvoice struct {
	parties   partySnapshot
	lines     []lineSnapshot
	compute   []computeLine
	vatExempt bool
}

// resolveScheduleInvoice gathers the data for a schedule-sourced invoice: the
// quote lines (name/unit/tax), the selected months' per-line HT (from cells),
// the resolved tax rates, and the frozen party block.
func (s *Server) resolveScheduleInvoice(ctx context.Context, userID, quoteID, scheduleID string, monthIndexes []int32) (*resolvedInvoice, int32, error) {
	q, lines, code, err := s.fetchQuoteAndLines(ctx, userID, quoteID)
	if err != nil || code != codes.Success {
		return nil, code, err
	}

	cellsResp, err := s.scheduleClient.GetScheduleCells(ctx, &scheduleGrpc.GetScheduleCellsRequest{
		ScheduleId: scheduleID,
		UserId:     userID,
	})
	if err != nil {
		return nil, codes.InternalError, fmt.Errorf("get schedule cells: %w", err)
	}
	if !cellsResp.GetSuccess() {
		return nil, codes.SourceNotEligible, nil
	}

	wantMonth := make(map[int32]struct{}, len(monthIndexes))
	for _, m := range monthIndexes {
		wantMonth[m] = struct{}{}
	}
	htByLine := make(map[string]int64)
	for _, c := range cellsResp.GetCells() {
		if _, ok := wantMonth[c.GetMonthIndex()]; ok {
			htByLine[c.GetQuoteLineId()] += c.GetAmountCents()
		}
	}

	return s.buildResolved(ctx, userID, q, lines, htByLine)
}

// resolveQuoteInvoice gathers the data for a whole-quote invoice: each line's
// full HT (unit_price × quantity).
func (s *Server) resolveQuoteInvoice(ctx context.Context, userID, quoteID string) (*resolvedInvoice, int32, error) {
	q, lines, code, err := s.fetchQuoteAndLines(ctx, userID, quoteID)
	if err != nil || code != codes.Success {
		return nil, code, err
	}

	htByLine := make(map[string]int64, len(lines))
	for _, l := range lines {
		htByLine[l.GetLineId()] = lineHTFromQuoteLine(l)
	}
	return s.buildResolved(ctx, userID, q, lines, htByLine)
}

// buildResolved resolves taxes, the issuer/client party block, and assembles
// the compute and snapshot line slices. Lines with zero billed HT are skipped.
func (s *Server) buildResolved(ctx context.Context, userID string, q *quoteGrpc.Quote, lines []*quoteGrpc.QuoteLine, htByLine map[string]int64) (*resolvedInvoice, int32, error) {
	taxCache := newTaxCache(s.usersClient)

	var parties partySnapshot
	var user *usersGrpc.User

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		resp, err := s.usersClient.GetUser(gctx, &usersGrpc.GetUserRequest{UserId: userID})
		if err != nil {
			return err
		}
		if !resp.GetSuccess() || resp.GetUser() == nil {
			return fmt.Errorf("get user: upstream %d", resp.GetCode())
		}
		user = resp.GetUser()
		return nil
	})

	var client *usersGrpc.Client
	g.Go(func() error {
		resp, err := s.usersClient.GetClient(gctx, &usersGrpc.GetClientRequest{ClientId: q.GetClientId(), UserId: userID})
		if err != nil {
			return err
		}
		if !resp.GetSuccess() || resp.GetClient() == nil {
			return fmt.Errorf("get client: upstream %d", resp.GetCode())
		}
		client = resp.GetClient()
		return nil
	})

	var clientAddr *usersGrpc.Address
	g.Go(func() error {
		resp, err := s.usersClient.GetAddress(gctx, &usersGrpc.GetAddressRequest{
			AddressId:  q.GetAddressId(),
			OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_CLIENT,
			OwnerId:    q.GetClientId(),
			AuthUserId: userID,
		})
		if err != nil {
			return err
		}
		if resp.GetSuccess() {
			clientAddr = resp.GetAddress()
		}
		return nil
	})

	var userAddr *usersGrpc.Address
	g.Go(func() error {
		if q.GetUserAddressId() == 0 {
			return nil
		}
		resp, err := s.usersClient.GetAddress(gctx, &usersGrpc.GetAddressRequest{
			AddressId:  q.GetUserAddressId(),
			OwnerType:  usersGrpc.OwnerType_OWNER_TYPE_USER,
			OwnerId:    userID,
			AuthUserId: userID,
		})
		if err != nil {
			return err
		}
		if resp.GetSuccess() {
			userAddr = resp.GetAddress()
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, codes.DependencyMissing, err
	}

	parties = buildPartySnapshot(user, userAddr, client, clientAddr)

	// OSS distance-selling: when the seller opted in and the client is a B2C
	// buyer in an EU country other than France, every line is taxed at the
	// destination country's VAT rate instead of the rate carried by the quote
	// line. oss is nil when OSS does not apply (normal domestic billing).
	oss, code, err := s.resolveOSSRate(ctx, user, parties)
	if err != nil || code != codes.Success {
		return nil, code, err
	}
	// Freeze whether OSS was applied, so the PDF can print the legal mention from
	// the snapshot alone (the seller's oss_enabled flag is mutable).
	parties.ossApplied = oss != nil

	// Stable ordering by quote-line position for the printed invoice.
	ordered := append([]*quoteGrpc.QuoteLine(nil), lines...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].GetPosition() < ordered[j].GetPosition() })

	var (
		snapLines []lineSnapshot
		compute   []computeLine
		pos       int32
	)
	for _, l := range ordered {
		ht, ok := htByLine[l.GetLineId()]
		if !ok || ht == 0 {
			continue // line not billed in this invoice (e.g. months not selected)
		}
		rate, label, code, err := taxCache.resolve(ctx, userID, l.GetTaxId())
		if err != nil {
			return nil, code, err
		}
		taxID := l.GetTaxId()
		if oss != nil {
			// Override the line's domestic rate with the destination rate. The
			// HT is unchanged; only the VAT rate/label differ. taxID is cleared
			// so the snapshot doesn't point at a domestic tax that wasn't applied.
			rate, label, taxID = oss.rate, oss.label, 0
		}
		numRate := parseRate(rate)
		snapLines = append(snapLines, lineSnapshot{
			position:       pos,
			quoteLineID:    l.GetLineId(),
			name:           l.GetName(),
			unit:           l.GetUnit(),
			quantity:       l.GetQuantity(),
			unitPriceCents: l.GetUnitPrice(),
			lineHTCents:    ht,
			taxID:          taxID,
			taxRate:        rate,
			taxLabel:       label,
		})
		compute = append(compute, computeLine{
			ht:        ht,
			taxID:     taxID,
			taxRate:   numRate,
			taxRateID: rate,
			taxLabel:  label,
		})
		pos++
	}

	return &resolvedInvoice{
		parties:   parties,
		lines:     snapLines,
		compute:   compute,
		vatExempt: strings.TrimSpace(user.GetVat()) == "",
	}, codes.Success, nil
}

// ossRate is the resolved destination-country VAT rate to apply to every line
// of an OSS distance-selling invoice.
type ossRate struct {
	rate  string // canonical rate string, e.g. "19"
	label string // tax label, e.g. "USt 19%"
}

// resolveOSSRate decides whether OSS applies and, if so, resolves the
// destination country's VAT rate. It returns (nil, Success, nil) when OSS does
// not apply (seller not opted in, B2B client, or client in FR / outside the EU).
// When OSS applies but no tax is configured for the client's country, it returns
// codes.OSSDestinationTaxMissing so emission is blocked rather than silently
// falling back to a domestic rate.
func (s *Server) resolveOSSRate(ctx context.Context, user *usersGrpc.User, parties partySnapshot) (*ossRate, int32, error) {
	if !user.GetOssEnabled() || parties.clientType != "individual" || parties.clientCountryID == 0 {
		return nil, codes.Success, nil
	}

	countryResp, err := s.usersClient.GetCountry(ctx, &usersGrpc.GetCountryRequest{CountryId: parties.clientCountryID})
	if err != nil {
		return nil, codes.InternalError, fmt.Errorf("get country: %w", err)
	}
	if !countryResp.GetSuccess() || countryResp.GetCountry() == nil {
		return nil, codes.DependencyMissing, nil
	}
	c := countryResp.GetCountry()
	if !c.GetIsEu() || c.GetCode() == "FR" {
		return nil, codes.Success, nil // domestic or non-EU: no OSS
	}

	taxesResp, err := s.usersClient.ListTaxesForCountry(ctx, &usersGrpc.ListTaxesForCountryRequest{CountryId: parties.clientCountryID})
	if err != nil {
		return nil, codes.InternalError, fmt.Errorf("list taxes for country: %w", err)
	}
	if !taxesResp.GetSuccess() {
		return nil, codes.InternalError, fmt.Errorf("list taxes for country: upstream %d", taxesResp.GetCode())
	}
	dest := pickDestinationTax(taxesResp.GetTaxes())
	if dest == nil {
		return nil, codes.OSSDestinationTaxMissing, nil
	}
	return &ossRate{rate: dest.GetRate(), label: dest.GetName()}, codes.Success, nil
}

// pickDestinationTax chooses the standard VAT rate for the destination country:
// the group's default tax if one is flagged, otherwise the highest rate (the
// standard rate is the highest among reduced rates). Returns nil when empty.
func pickDestinationTax(taxes []*usersGrpc.Tax) *usersGrpc.Tax {
	var best *usersGrpc.Tax
	for _, t := range taxes {
		if t.GetIsDefault() {
			return t
		}
		if best == nil || parseRate(t.GetRate()) > parseRate(best.GetRate()) {
			best = t
		}
	}
	return best
}

// fetchQuoteAndLines loads the quote and its lines, returning a business code on
// any upstream failure.
func (s *Server) fetchQuoteAndLines(ctx context.Context, userID, quoteID string) (*quoteGrpc.Quote, []*quoteGrpc.QuoteLine, int32, error) {
	resp, err := s.quoteClient.GetQuote(ctx, &quoteGrpc.GetQuoteRequest{QuoteId: quoteID, UserId: userID})
	if err != nil {
		return nil, nil, codes.InternalError, fmt.Errorf("get quote: %w", err)
	}
	if !resp.GetSuccess() || resp.GetQuote() == nil {
		return nil, nil, codes.NotFound, nil
	}
	return resp.GetQuote(), resp.GetLines(), codes.Success, nil
}

func lineHTFromQuoteLine(l *quoteGrpc.QuoteLine) int64 {
	qty, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(l.GetQuantity()), ",", "."), 64)
	if err != nil {
		return 0
	}
	return int64(float64(l.GetUnitPrice())*qty + 0.5)
}

// clientTypeToString maps the users-service ClientType enum to the DB string
// frozen in the party snapshot. UNSPECIFIED maps to "" so a not-yet-classified
// client doesn't fabricate a B2C/B2B nature on the invoice.
func clientTypeToString(t usersGrpc.ClientType) string {
	switch t {
	case usersGrpc.ClientType_CLIENT_TYPE_INDIVIDUAL:
		return "individual"
	case usersGrpc.ClientType_CLIENT_TYPE_BUSINESS:
		return "business"
	default:
		return ""
	}
}

func buildPartySnapshot(user *usersGrpc.User, userAddr *usersGrpc.Address, client *usersGrpc.Client, clientAddr *usersGrpc.Address) partySnapshot {
	p := partySnapshot{}
	if user != nil {
		p.issuerCompany = user.GetCompany()
		p.issuerSiren = user.GetSiren()
		p.issuerVat = user.GetVat()
		p.issuerEmail = user.GetEmail()
		p.issuerPhone = user.GetPhone()
		p.issuerLogoURL = user.GetLogoUrl()
	}
	if userAddr != nil {
		p.issuerStreet = userAddr.GetStreet()
		p.issuerAdditional = userAddr.GetAdditionalStreet()
		p.issuerZip = userAddr.GetZipCode()
		p.issuerCity = userAddr.GetCity()
	}
	if client != nil {
		p.clientFirstName = client.GetFirstName()
		p.clientLastName = client.GetLastName()
		p.clientCompany = client.GetCompany()
		p.clientSiren = client.GetSiren()
		p.clientVat = client.GetVat()
		p.clientEmail = client.GetEmail()
		p.clientType = clientTypeToString(client.GetClientType())
	}
	if clientAddr != nil {
		p.clientStreet = clientAddr.GetStreet()
		p.clientAdditional = clientAddr.GetAdditionalStreet()
		p.clientZip = clientAddr.GetZipCode()
		p.clientCity = clientAddr.GetCity()
		p.clientCountryID = clientAddr.GetCountryId()
	}
	return p
}

// taxCache resolves tax_id → (rate, label) once per id within a request.
type taxCache struct {
	client usersGrpc.UserServiceClient
	mu     sync.Mutex
	cache  map[int32]taxEntry
}

type taxEntry struct {
	rate  string
	label string
}

func newTaxCache(client usersGrpc.UserServiceClient) *taxCache {
	return &taxCache{client: client, cache: make(map[int32]taxEntry)}
}

func (c *taxCache) resolve(ctx context.Context, userID string, taxID int32) (rate, label string, code int32, err error) {
	if taxID == 0 {
		return "0", "", codes.Success, nil
	}
	c.mu.Lock()
	if e, ok := c.cache[taxID]; ok {
		c.mu.Unlock()
		return e.rate, e.label, codes.Success, nil
	}
	c.mu.Unlock()

	resp, err := c.client.GetTax(ctx, &usersGrpc.GetTaxRequest{TaxId: taxID})
	if err != nil {
		return "", "", codes.InternalError, fmt.Errorf("get tax %d: %w", taxID, err)
	}
	if !resp.GetSuccess() || resp.GetTax() == nil {
		return "", "", codes.DependencyMissing, nil
	}
	e := taxEntry{rate: resp.GetTax().GetRate(), label: resp.GetTax().GetName()}
	c.mu.Lock()
	c.cache[taxID] = e
	c.mu.Unlock()
	return e.rate, e.label, codes.Success, nil
}
