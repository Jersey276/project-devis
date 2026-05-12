package quote

import (
	"context"
	"fmt"
	"strings"
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
	upstreamNotFound      int32 = 1001
	upstreamInvalidInput  int32 = 1003
	upstreamInternalError int32 = 2001
)

func Export(ctx context.Context, qc quote.QuoteServiceClient, uc users.UserServiceClient,
	gt *gotenberg.Client, req *exportGrpc.ExportQuoteRequest) (*exportGrpc.ExportQuoteResponse, error) {

	if req.QuoteId == "" || req.UserId == "" {
		return fail(codes.InvalidInput), nil
	}

	qResp, err := qc.GetQuote(ctx, &quote.GetQuoteRequest{QuoteId: req.QuoteId, UserId: req.UserId})
	if err != nil {
		return fail(codes.InternalError), err
	}
	if !qResp.Success || qResp.Quote == nil {
		return fail(mapUpstreamCode(qResp.Code)), nil
	}

	q := qResp.Quote
	lines := qResp.Lines

	var (
		user        *users.User
		userAddr    *users.Address
		client      *users.Client
		clientAddr  *users.Address
	)

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		resp, err := uc.GetUser(gctx, &users.GetUserRequest{UserId: req.UserId})
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		if !resp.Success || resp.User == nil {
			return fmt.Errorf("get user: upstream code %d", resp.Code)
		}
		user = resp.User
		return nil
	})

	// Sender address: first address (lowest id) owned by the authenticated user.
	// Optional — render falls back gracefully if the user has no address yet.
	g.Go(func() error {
		resp, err := uc.ListAddresses(gctx, &users.ListAddressesRequest{
			OwnerType:  users.OwnerType_OWNER_TYPE_USER,
			OwnerId:    req.UserId,
			AuthUserId: req.UserId,
		})
		if err != nil {
			return fmt.Errorf("list user addresses: %w", err)
		}
		if !resp.Success {
			return fmt.Errorf("list user addresses: upstream code %d", resp.Code)
		}
		if len(resp.Addresses) > 0 {
			userAddr = resp.Addresses[0]
		}
		return nil
	})

	g.Go(func() error {
		resp, err := uc.GetClient(gctx, &users.GetClientRequest{ClientId: q.ClientId, UserId: req.UserId})
		if err != nil {
			return fmt.Errorf("get client: %w", err)
		}
		if !resp.Success || resp.Client == nil {
			return fmt.Errorf("get client: upstream code %d", resp.Code)
		}
		client = resp.Client
		return nil
	})

	g.Go(func() error {
		resp, err := uc.GetAddress(gctx, &users.GetAddressRequest{
			AddressId:  q.AddressId,
			OwnerType:  users.OwnerType_OWNER_TYPE_CLIENT,
			OwnerId:    q.ClientId,
			AuthUserId: req.UserId,
		})
		if err != nil {
			return fmt.Errorf("get address: %w", err)
		}
		if !resp.Success || resp.Address == nil {
			return fmt.Errorf("get address: upstream code %d", resp.Code)
		}
		clientAddr = resp.Address
		return nil
	})

	if err := g.Wait(); err != nil {
		return fail(codes.InternalError), err
	}

	pdfBytes, err := Render(ctx, gt, q, lines, user, userAddr, client, clientAddr)
	if err != nil {
		return fail(codes.InternalError), err
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

// Output is UTF-8; callers must encode it for the transport
// (Content-Disposition uses RFC 5987 filename*).
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
