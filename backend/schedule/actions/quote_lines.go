package actions

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	quoteGrpc "project-devis-schedule/services/quotegrpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var quoteLineExpectedCentsFetcher = fetchQuoteLineExpectedCentsFromQuoteService

// SetQuoteLineExpectedCentsFetcherForTests allows tests to stub quote lookups.
func SetQuoteLineExpectedCentsFetcherForTests(fetcher func(context.Context, string, string) (map[string]int64, error)) func() {
	previous := quoteLineExpectedCentsFetcher
	if fetcher == nil {
		quoteLineExpectedCentsFetcher = fetchQuoteLineExpectedCentsFromQuoteService
	} else {
		quoteLineExpectedCentsFetcher = fetcher
	}
	return func() {
		quoteLineExpectedCentsFetcher = previous
	}
}

func getQuoteLineExpectedCents(ctx context.Context, userID, quoteID string) (map[string]int64, error) {
	return quoteLineExpectedCentsFetcher(ctx, userID, quoteID)
}

func fetchQuoteLineExpectedCentsFromQuoteService(ctx context.Context, userID, quoteID string) (map[string]int64, error) {
	address := os.Getenv("QUOTE_SERVICE_ADDRESS")
	if strings.TrimSpace(address) == "" {
		address = "localhost:50053"
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("connect quote grpc: %w", err)
	}
	defer conn.Close()

	client := quoteGrpc.NewQuoteServiceClient(conn)
	resp, err := client.ListQuoteLines(ctx, &quoteGrpc.ListQuoteLinesRequest{QuoteId: quoteID, UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("list quote lines: %w", err)
	}
	if !resp.GetSuccess() {
		return map[string]int64{}, nil
	}

	amounts := make(map[string]int64, len(resp.GetLines()))
	for _, line := range resp.GetLines() {
		lineID := strings.TrimSpace(line.GetLineId())
		if lineID == "" {
			continue
		}
		qty, err := strconv.ParseFloat(strings.TrimSpace(line.GetQuantity()), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid quote line quantity for %s: %w", lineID, err)
		}
		// quote unit_price is stored in cents.
		amounts[lineID] = int64(float64(line.GetUnitPrice())*qty + 0.5)
	}

	return amounts, nil
}
