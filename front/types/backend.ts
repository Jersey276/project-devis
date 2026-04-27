export type BackendQuote = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  created_at: string;
  updated_at: string;
};

export type BackendQuoteLineType = "simple" | "multiple";

export type BackendQuoteLine = {
  line_id: string;
  quote_id: string;
  type: BackendQuoteLineType;
  name: string;
  quantity: string;
  unit: string;
  unit_price: number;
  data: Record<string, unknown>;
  position: number;
};

export type QuoteListStatus = "brouillon" | "archivé";

export function quoteListStatus(quote: BackendQuote): QuoteListStatus {
  return quote.archived_at ? "archivé" : "brouillon";
}
