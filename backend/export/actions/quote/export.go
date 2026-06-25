package quote

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/errgroup"

	"project-devis-export/actions/codes"
	"project-devis-export/internal/slug"
	"project-devis-export/quote"
	exportGrpc "project-devis-export/services/grpc"
	"project-devis-export/users"
)

// Upstream codes (kept in sync with backend/quote/actions/codes/codes.go and
// backend/users/actions/codes/codes.go). Mirrored here so we don't import
// across services.
const (
	upstreamNotFound     int32 = 1001
	upstreamInvalidInput int32 = 1003
)

// pdfConverter is the slice of *gotenberg.Client we depend on; defined here so
// tests can substitute an in-memory implementation.
type pdfConverter interface {
	Convert(ctx context.Context, html []byte) ([]byte, error)
}

func Export(ctx context.Context, qc quote.QuoteServiceClient, uc users.UserServiceClient,
	gt pdfConverter, req *exportGrpc.ExportQuoteRequest) (*exportGrpc.ExportQuoteResponse, error) {

	if req.QuoteId == "" || req.UserId == "" {
		return fail(codes.InvalidInput), nil
	}

	// Phase 1: GetQuote and GetUser have no inter-dependency. The user
	// address fetch moved to Phase 2 because it now keys off the quote's
	// user_address_id (the prestataire address chosen at quote creation).
	var (
		qResp         *quote.GetQuoteResponse
		user          *users.User
		firstUpstream atomic.Int32 // first non-zero local code recorded by a goroutine
	)
	recordCode := func(local int32) {
		firstUpstream.CompareAndSwap(0, local)
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := qc.GetQuote(gctx, &quote.GetQuoteRequest{QuoteId: req.QuoteId, UserId: req.UserId})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Quote == nil {
			recordCode(mapQuoteCode(resp.Code))
			return fmt.Errorf("get quote: upstream %d", resp.Code)
		}
		qResp = resp
		return nil
	})

	g.Go(func() error {
		resp, err := uc.GetUser(gctx, &users.GetUserRequest{UserId: req.UserId})
		if err != nil {
			return err
		}
		if !resp.Success || resp.User == nil {
			// The auth middleware already resolved this user; missing here means
			// upstream data inconsistency, not a 404 we should leak to clients.
			recordCode(codes.InternalError)
			return fmt.Errorf("get user: upstream %d", resp.Code)
		}
		user = resp.User
		return nil
	})

	if err := g.Wait(); err != nil {
		if code := firstUpstream.Load(); code != 0 {
			return fail(code), nil
		}
		return nil, err
	}

	// Phase 2: GetClient, GetAddress (client side) and GetAddress (user side)
	// all need ids from the quote.
	q := qResp.Quote
	if q.State == quote.QuoteState_QUOTE_STATE_DROP {
		return fail(codes.QuoteRefused), nil
	}
	lines := qResp.Lines
	var (
		client     *users.Client
		clientAddr *users.Address
		userAddr   *users.Address
	)

	g2, g2ctx := errgroup.WithContext(ctx)

	g2.Go(func() error {
		resp, err := uc.GetClient(g2ctx, &users.GetClientRequest{ClientId: q.ClientId, UserId: req.UserId})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Client == nil {
			recordCode(mapDependencyCode(resp.Code))
			return fmt.Errorf("get client: upstream %d", resp.Code)
		}
		client = resp.Client
		return nil
	})

	g2.Go(func() error {
		resp, err := uc.GetAddress(g2ctx, &users.GetAddressRequest{
			AddressId:  q.AddressId,
			OwnerType:  users.OwnerType_OWNER_TYPE_CLIENT,
			OwnerId:    q.ClientId,
			AuthUserId: req.UserId,
		})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Address == nil {
			recordCode(mapDependencyCode(resp.Code))
			return fmt.Errorf("get address: upstream %d", resp.Code)
		}
		clientAddr = resp.Address
		return nil
	})

	g2.Go(func() error {
		// Newer quotes always carry a user_address_id (DB column is NOT NULL).
		// Legacy paths that called Export without one would land here with 0;
		// keep a list-based fallback so the PDF still renders.
		if q.UserAddressId == 0 {
			resp, err := uc.ListAddresses(g2ctx, &users.ListAddressesRequest{
				OwnerType:  users.OwnerType_OWNER_TYPE_USER,
				OwnerId:    req.UserId,
				AuthUserId: req.UserId,
			})
			if err != nil {
				return err
			}
			if !resp.Success {
				recordCode(codes.InternalError)
				return fmt.Errorf("list user addresses: upstream %d", resp.Code)
			}
			if len(resp.Addresses) > 0 {
				userAddr = resp.Addresses[0]
			}
			return nil
		}
		resp, err := uc.GetAddress(g2ctx, &users.GetAddressRequest{
			AddressId:  q.UserAddressId,
			OwnerType:  users.OwnerType_OWNER_TYPE_USER,
			OwnerId:    req.UserId,
			AuthUserId: req.UserId,
		})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Address == nil {
			recordCode(mapDependencyCode(resp.Code))
			return fmt.Errorf("get user address: upstream %d", resp.Code)
		}
		userAddr = resp.Address
		return nil
	})

	if err := g2.Wait(); err != nil {
		if code := firstUpstream.Load(); code != 0 {
			return fail(code), nil
		}
		return nil, err
	}

	// Phase 3: resolve distinct tax_ids from quote lines.
	taxes := make(map[int32]*users.Tax)
	seen := map[int32]bool{}
	var taxIDs []int32
	for _, l := range lines {
		if l.TaxId != 0 && !seen[l.TaxId] {
			seen[l.TaxId] = true
			taxIDs = append(taxIDs, l.TaxId)
		}
	}
	if len(taxIDs) > 0 {
		var mu sync.Mutex
		g3, g3ctx := errgroup.WithContext(ctx)
		for _, id := range taxIDs {
			id := id
			g3.Go(func() error {
				resp, err := uc.GetTax(g3ctx, &users.GetTaxRequest{TaxId: id})
				if err != nil {
					return err
				}
				if resp.Success && resp.Tax != nil {
					mu.Lock()
					taxes[id] = resp.Tax
					mu.Unlock()
				}
				return nil
			})
		}
		if err := g3.Wait(); err != nil {
			return nil, err
		}
	}

	pdfBytes, err := Render(ctx, gt, renderInput{
		Quote:         q,
		Lines:         lines,
		User:          user,
		UserAddress:   userAddr,
		Client:        client,
		ClientAddress: clientAddr,
		Taxes:         taxes,
	})
	if err != nil {
		return nil, err
	}

	return &exportGrpc.ExportQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		Pdf:      pdfBytes,
		Filename: buildFilename(q),
	}, nil
}

func buildFilename(q *quote.Quote) string {
	s := slug.Slugify(q.Name)
	if s == "" {
		return fmt.Sprintf("devis-%s.pdf", q.QuoteId)
	}
	return fmt.Sprintf("devis-%s.pdf", s)
}

func fail(code int32) *exportGrpc.ExportQuoteResponse {
	return &exportGrpc.ExportQuoteResponse{Success: false, Code: code}
}

// mapQuoteCode translates an upstream quote-service code into the local code
// for the requested quote itself (NotFound means "no such quote").
func mapQuoteCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.NotFound
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}

// mapDependencyCode translates an upstream users-service code raised for a
// referenced entity (client/address). NotFound here means a *dependency* is
// missing — distinct from the quote being missing.
func mapDependencyCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.DependencyMissing
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}
