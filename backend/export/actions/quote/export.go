package quote

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"unicode"

	"golang.org/x/sync/errgroup"

	"project-devis-export/actions/codes"
	"project-devis-export/quote"
	"project-devis-export/services/gotenberg"
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

func Export(ctx context.Context, qc quote.QuoteServiceClient, uc users.UserServiceClient,
	gt *gotenberg.Client, req *exportGrpc.ExportQuoteRequest) (*exportGrpc.ExportQuoteResponse, error) {

	if req.QuoteId == "" || req.UserId == "" {
		return fail(codes.InvalidInput), nil
	}

	// Phase 1: GetQuote, GetUser, and the user's address list have no
	// inter-dependency, so we fan them out together.
	var (
		qResp    *quote.GetQuoteResponse
		user     *users.User
		userAddr *users.Address
		firstUpstream atomic.Int32 // first soft (resp.Success=false) code, mapped to local
	)
	recordUpstream := func(c int32) {
		firstUpstream.CompareAndSwap(0, mapUpstreamCode(c))
	}

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := qc.GetQuote(gctx, &quote.GetQuoteRequest{QuoteId: req.QuoteId, UserId: req.UserId})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Quote == nil {
			recordUpstream(resp.Code)
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
			recordUpstream(resp.Code)
			return fmt.Errorf("get user: upstream %d", resp.Code)
		}
		user = resp.User
		return nil
	})

	g.Go(func() error {
		resp, err := uc.ListAddresses(gctx, &users.ListAddressesRequest{
			OwnerType:  users.OwnerType_OWNER_TYPE_USER,
			OwnerId:    req.UserId,
			AuthUserId: req.UserId,
		})
		if err != nil {
			return err
		}
		if !resp.Success {
			recordUpstream(resp.Code)
			return fmt.Errorf("list user addresses: upstream %d", resp.Code)
		}
		if len(resp.Addresses) > 0 {
			userAddr = resp.Addresses[0]
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		if code := firstUpstream.Load(); code != 0 {
			return fail(code), nil
		}
		return nil, err
	}

	// Phase 2: GetClient and GetAddress both need ClientId from the quote.
	q := qResp.Quote
	lines := qResp.Lines
	var (
		client     *users.Client
		clientAddr *users.Address
	)

	g2, g2ctx := errgroup.WithContext(ctx)

	g2.Go(func() error {
		resp, err := uc.GetClient(g2ctx, &users.GetClientRequest{ClientId: q.ClientId, UserId: req.UserId})
		if err != nil {
			return err
		}
		if !resp.Success || resp.Client == nil {
			recordUpstream(resp.Code)
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
			recordUpstream(resp.Code)
			return fmt.Errorf("get address: upstream %d", resp.Code)
		}
		clientAddr = resp.Address
		return nil
	})

	if err := g2.Wait(); err != nil {
		if code := firstUpstream.Load(); code != 0 {
			return fail(code), nil
		}
		return nil, err
	}

	pdfBytes, err := Render(ctx, gt, renderInput{
		Quote:         q,
		Lines:         lines,
		User:          user,
		UserAddress:   userAddr,
		Client:        client,
		ClientAddress: clientAddr,
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
	slug := slugify(q.Name)
	if slug == "" {
		return fmt.Sprintf("devis-%s.pdf", q.QuoteId)
	}
	return fmt.Sprintf("devis-%s.pdf", slug)
}

func slugify(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			prevDash = false
		case unicode.IsSpace(r) || r == '_' || r == '-':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	out := b.String()
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	return out
}

func fail(code int32) *exportGrpc.ExportQuoteResponse {
	return &exportGrpc.ExportQuoteResponse{Success: false, Code: code}
}

func mapUpstreamCode(c int32) int32 {
	switch c {
	case upstreamNotFound:
		return codes.NotFound
	case upstreamInvalidInput:
		return codes.InvalidInput
	default:
		return codes.InternalError
	}
}
