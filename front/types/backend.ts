export type BackendQuoteState = "draft" | "sent" | "validated" | "drop";

export type BackendQuote = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  state: BackendQuoteState;
  client_id: string;
  address_id: number;
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
  tax_id: number | null;
};

export type BackendTax = {
  id: number;
  name: string;
  rate: string;
  country_group_id: number;
  is_default: boolean;
  original_tax_id?: number;
  version?: number;
  superseded_at?: string;
  superseded_by?: number;
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

export type BackendClient = {
  client_id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  company: string;
  siren: string;
  vat: string;
  archived: boolean;
};

export type BackendAddressOwnerType = "user" | "client";

export type BackendAddress = {
  id: number;
  owner_type: BackendAddressOwnerType;
  owner_id: string;
  name: string;
  street: string;
  additional_street: string;
  city: string;
  zip_code: string;
  country_id: number;
  email: string;
  phone: string;
  archived: boolean;
};
