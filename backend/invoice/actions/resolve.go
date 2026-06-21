package actions

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"project-devis-invoice/actions/codes"
	quoteGrpc "project-devis-invoice/services/quotegrpc"
	scheduleGrpc "project-devis-invoice/services/schedulegrpc"
	usersGrpc "project-devis-invoice/services/usersgrpc"
)

type partySnapshot struct {
	issuerCompany         string
	issuerSiren           string
	issuerSiret           string
	issuerVat             string
	issuerEmail           string
	issuerPhone           string
	issuerLogoURL         string
	issuerStreet          string
	issuerAdditional      string
	issuerZip             string
	issuerCity            string
	issuerIban            string
	issuerBic             string
	clientFirstName       string
	clientLastName        string
	clientCompany         string
	clientSiren           string
	clientSiret           string
	clientVat             string
	clientEmail           string
	clientStreet          string
	clientAdditional      string
	clientZip             string
	clientCity            string
	clientType            string
	clientCountryID       int32
	ossApplied            bool
	countsTowardThreshold bool
	issuerCountryCode     string
	clientCountryCode     string
}

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

type resolvedInvoice struct {
	parties   partySnapshot
	lines     []lineSnapshot
	compute   []computeLine
	vatExempt bool
}

func (s *Server) resolveScheduleInvoice(ctx context.Context, invoiceID, userID, quoteID, scheduleID string, monthIndexes []int32, issuedAt time.Time) (*resolvedInvoice, int32, error) {
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

	return s.buildResolved(ctx, invoiceID, userID, q, lines, htByLine, issuedAt)
}

func (s *Server) resolveQuoteInvoice(ctx context.Context, invoiceID, userID, quoteID string, issuedAt time.Time) (*resolvedInvoice, int32, error) {
	q, lines, code, err := s.fetchQuoteAndLines(ctx, userID, quoteID)
	if err != nil || code != codes.Success {
		return nil, code, err
	}

	htByLine := make(map[string]int64, len(lines))
	for _, l := range lines {
		htByLine[l.GetLineId()] = lineHTFromQuoteLine(l)
	}
	return s.buildResolved(ctx, invoiceID, userID, q, lines, htByLine, issuedAt)
}

func (s *Server) buildResolved(ctx context.Context, invoiceID, userID string, q *quoteGrpc.Quote, lines []*quoteGrpc.QuoteLine, htByLine map[string]int64, issuedAt time.Time) (*resolvedInvoice, int32, error) {
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

	var clientCountry *usersGrpc.Country
	if parties.clientCountryID != 0 {
		c, code, err := s.resolveCountry(ctx, parties.clientCountryID)
		if err != nil || code != codes.Success {
			return nil, code, err
		}
		clientCountry = c
		parties.clientCountryCode = clientCountry.GetCode()
	}
	if issuerCountryID := userAddr.GetCountryId(); issuerCountryID != 0 {
		issuerCountry, code, err := s.resolveCountry(ctx, issuerCountryID)
		if err != nil || code != codes.Success {
			return nil, code, err
		}
		parties.issuerCountryCode = issuerCountry.GetCode()
	}

	parties.countsTowardThreshold = isIntraEUB2C(parties.clientType, clientCountry)
	cumulativeHTCents, err := s.ossCumulativeHTForYear(ctx, userID, invoiceID, issuedAt)
	if err != nil {
		return nil, codes.InternalError, err
	}
	priorYearOver, _, err := s.ossPriorYearOverThreshold(ctx, userID, issuedAt)
	if err != nil {
		return nil, codes.InternalError, err
	}

	oss, code, err := s.resolveOSSRate(ctx, user, cumulativeHTCents, priorYearOver, parties, clientCountry)
	if err != nil || code != codes.Success {
		return nil, code, err
	}

	parties.ossApplied = oss != nil

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
			continue
		}
		rate, label, code, err := taxCache.resolve(ctx, userID, l.GetTaxId())
		if err != nil {
			return nil, code, err
		}
		taxID := l.GetTaxId()
		if oss != nil {

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

type ossRate struct {
	rate  string
	label string
}

func (s *Server) resolveCountry(ctx context.Context, countryID int32) (*usersGrpc.Country, int32, error) {
	resp, err := s.usersClient.GetCountry(ctx, &usersGrpc.GetCountryRequest{CountryId: countryID})
	if err != nil {
		return nil, codes.InternalError, fmt.Errorf("get country: %w", err)
	}
	if !resp.GetSuccess() || resp.GetCountry() == nil {
		return nil, codes.DependencyMissing, nil
	}
	return resp.GetCountry(), codes.Success, nil
}

func (s *Server) resolveOSSRate(ctx context.Context, user *usersGrpc.User, cumulativeHTCents int64, priorYearOverThreshold bool, parties partySnapshot, clientCountry *usersGrpc.Country) (*ossRate, int32, error) {
	if !ossApplies(user.GetOssEnabled(), cumulativeHTCents, priorYearOverThreshold, parties.clientType, clientCountry) {
		return nil, codes.Success, nil
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
		p.issuerSiret = user.GetSiret()
		p.issuerVat = user.GetVat()
		p.issuerEmail = user.GetEmail()
		p.issuerPhone = user.GetPhone()
		p.issuerLogoURL = user.GetLogoUrl()
		p.issuerIban = user.GetIban()
		p.issuerBic = user.GetBic()
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
		p.clientSiret = client.GetSiret()
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
