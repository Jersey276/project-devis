export type BackendQuoteState = "draft" | "sent" | "validated" | "drop";

export type BackendQuote = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  state: BackendQuoteState;
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

export const QUOTE_STATE_LABEL: Record<BackendQuoteState, string> = {
  draft: "Brouillon",
  sent: "Envoyé",
  validated: "Validé",
  drop: "Abandonné",
};

export type QuoteListStatus =
  | "brouillon"
  | "envoyé"
  | "validé"
  | "abandonné"
  | "archivé";

export function quoteListStatus(quote: BackendQuote): QuoteListStatus {
  if (quote.archived_at) return "archivé";
  const label = QUOTE_STATE_LABEL[quote.state];
  return (label ? label.toLowerCase() : "brouillon") as QuoteListStatus;
}
