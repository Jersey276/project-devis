export type BackendQuoteState = "draft" | "sent" | "validated" | "drop";

export type BackendQuote = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  state: BackendQuoteState;
  client_id: string;
  address_id: number;
  user_address_id: number;
  created_at: string;
  updated_at: string;
  // Present on GET /api/quotes only — TTC total in cents, aggregated by the gateway.
  total_ttc?: number;
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

export type QuoteListState = BackendQuoteState | "archived";

export function quoteListState(quote: BackendQuote): QuoteListState {
  if (quote.archived_at) return "archived";
  return quote.state ?? "draft";
}

export type BackendTemplateType = "quote_document" | "quote_line" | "document_design";
export type BackendTemplateTargetResource = "quote" | "invoice" | "schedule";

export type BackendTemplate = {
  template_id: string;
  user_id: string;
  template_type: BackendTemplateType;
  target_resource: BackendTemplateTargetResource;
  name: string;
  archived_at: string | null;
  payload_version: number;
  payload: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type BackendTemplateLine = {
  line_id: string;
  template_id: string;
  type: string;
  name: string;
  quantity: string;
  unit: string | null;
  unit_price: number;
  data: Record<string, unknown>;
  position: number;
  tax_id: number | null;
  created_at: string;
  updated_at: string;
};

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
