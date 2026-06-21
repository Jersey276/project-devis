export type BackendQuoteState =
  | "draft"
  | "negociation"
  | "validated"
  | "drop";

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

export type QuoteLineKind =
  | "line"
  | "text"
  | "group"
  | "detailed"
  | "subline"
  | "fee";

export type QuoteLineData = {
  kind?: QuoteLineKind;
  description?: string;
  option?: boolean;
  parent_line_id?: string;
  /** Set on a top-level fee line (kind="fee"): the catalog entry it mirrors. */
  fee_id?: string;
  sublines?: Array<{
    name: string;
    quantity: string;
    unit?: string;
    unit_price: number;
    option?: boolean;
    /** Set when the subline was added from a fee catalog entry. */
    fee_id?: string;
    /** Frontend-only stable React key — stripped before API calls. */
    _key?: string;
  }>;
};

export type BackendQuoteLine = {
  line_id: string;
  quote_id: string;
  type: BackendQuoteLineType;
  name: string;
  quantity: string;
  unit: string;
  unit_price: number;
  data: QuoteLineData;
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

export type FeeCategory = "fixed" | "service";

export type BackendFee = {
  fee_id: string;
  category: FeeCategory;
  name: string;
  unit: string;
  unit_price: number;
  tax_id: number | null;
  archived: boolean;
};

export type QuoteListState = BackendQuoteState | "archived";

export function quoteListState(quote: BackendQuote): QuoteListState {
  if (quote.archived_at) return "archived";
  return quote.state ?? "draft";
}

export type BackendTemplateType =
  | "quote_document"
  | "quote_line"
  | "document_design";
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
  data: QuoteLineData;
  position: number;
  tax_id: number | null;
  created_at: string;
  updated_at: string;
};

// "individual" = B2C, "business" = B2B. Empty string for legacy clients not yet
// classified.
export type ClientType = "individual" | "business" | "";

export type BackendClient = {
  client_id: string;
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  company: string;
  siren: string;
  siret: string;
  vat: string;
  archived: boolean;
  client_type: ClientType;
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

export type BackendScheduleStatus = "DRAFT" | "NEGOCIATE" | "DENIED" | "VALID";

export type BackendScheduleSummary = {
  schedule_id: string;
  quote_id: string;
  status: BackendScheduleStatus;
  name: string;
  start_month: string;
  duration_months: number;
};

export type BackendScheduleLineSummary = {
  quote_line_id: string;
  planned_cents: number;
  expected_cents: number;
};

export type BackendScheduleColumnTotal = {
  month_index: number;
  amount_cents: number;
};

export type BackendScheduleCell = {
  quote_line_id: string;
  month_index: number;
  amount_cents: number;
};

export type BackendScheduleDetails = {
  schedule_id: string;
  quote_id: string;
  status: BackendScheduleStatus;
  name: string;
  start_month: string;
  duration_months: number;
  lines: BackendScheduleLineSummary[];
  cells?: BackendScheduleCell[];
  column_totals: BackendScheduleColumnTotal[];
  quote_total_cents: number;
  planned_total_cents: number;
};

// ─── Invoice ─────────────────────────────────────────────────────────────────

export type BackendInvoiceStatus = "DRAFT" | "ISSUED" | "PAID" | "CANCELLED";

// E-invoicing lifecycle status (réforme FR B2B), orthogonal to the business
// status. "NONE" = no platform lifecycle yet.
export type BackendInvoiceLifecycleStatus =
  | "NONE"
  | "DEPOSITED"
  | "RECEIVED"
  | "APPROVED"
  | "REJECTED"
  | "COLLECTED";

export type BackendInvoiceLifecycleEvent = {
  status: Exclude<BackendInvoiceLifecycleStatus, "NONE">;
  note: string;
  created_at: string;
};

export type BackendInvoiceSummary = {
  invoice_id: string;
  invoice_number: string;
  status: BackendInvoiceStatus;
  quote_id: string;
  schedule_id: string;
  issued_at: string;
  due_date: string;
  total_ttc_cents: number;
  lifecycle_status: BackendInvoiceLifecycleStatus;
};

export type BackendOSSThresholdStatus = {
  year: number;
  cumulative_ht_cents: number;
  threshold_cents: number;
  oss_enabled: boolean;
  oss_active: boolean;
  // N-1 rule (art. 259 D CGI): prior year crossed the threshold → destination
  // VAT from the first euro this year.
  prior_year_over_threshold: boolean;
  prior_year_cumulative_ht_cents: number;
};

export type BackendInvoiceParty = {
  company: string;
  first_name: string;
  last_name: string;
  siren: string;
  siret: string;
  vat: string;
  email: string;
  phone: string;
  logo_url: string;
  street: string;
  additional_street: string;
  zip_code: string;
  city: string;
  iban: string;
  bic: string;
};

export type BackendInvoiceLine = {
  quote_line_id: string;
  name: string;
  unit: string;
  quantity: string;
  unit_price_cents: number;
  line_ht_cents: number;
  tax_id: number;
  tax_rate: string;
  tax_label: string;
};

export type BackendInvoiceVatLine = {
  tax_rate: string;
  base_ht_cents: number;
  vat_cents: number;
};

export type BackendInvoiceDetails = {
  invoice_id: string;
  quote_id: string;
  schedule_id: string;
  billed_month_indexes: number[];
  status: BackendInvoiceStatus;
  invoice_number: string;
  issued_at: string;
  sale_date: string;
  due_date: string;
  issuer: BackendInvoiceParty;
  client: BackendInvoiceParty;
  lines: BackendInvoiceLine[];
  vat_breakdown: BackendInvoiceVatLine[];
  total_ht_cents: number;
  total_vat_cents: number;
  total_ttc_cents: number;
  vat_exempt: boolean;
  credited_positions: number[];
  lifecycle_status: BackendInvoiceLifecycleStatus;
};

export type BackendCreditNoteSummary = {
  credit_note_id: string;
  credit_note_number: string;
  invoice_id: string;
  invoice_number: string;
  issued_at: string;
  is_total: boolean;
  total_ttc_cents: number;
};

export type BackendCreditNoteDetails = {
  credit_note_id: string;
  invoice_id: string;
  invoice_number: string;
  credit_note_number: string;
  issued_at: string;
  reason: string;
  is_total: boolean;
  issuer: BackendInvoiceParty;
  client: BackendInvoiceParty;
  lines: BackendInvoiceLine[];
  vat_breakdown: BackendInvoiceVatLine[];
  total_ht_cents: number;
  total_vat_cents: number;
  total_ttc_cents: number;
  vat_exempt: boolean;
};

export type ScheduleBalanceState = "under" | "balanced" | "over";

export function scheduleBalanceState(
  plannedCents: number,
  expectedCents: number,
): ScheduleBalanceState {
  if (plannedCents < expectedCents) return "under";
  if (plannedCents > expectedCents) return "over";
  return "balanced";
}

export type SubscriptionTier = "free" | "pro" | "enterprise";

export type BackendPlan = {
  plan_id: number;
  name: string;
  tier: SubscriptionTier;
  price_cents: number;
  billing_cycle: "monthly" | "yearly" | "none";
  features: Record<string, number>;
  active: boolean;
  stripe_price_id?: string | null;
};

export type BackendSubscription = {
  subscription_id: string;
  user_id: string;
  plan_id: number;
  tier: SubscriptionTier;
  status: "active" | "cancelled" | "expired";
  current_period_start: string;
  current_period_end: string | null;
  cancel_at_period_end: boolean;
  stripe_subscription_id: string | null;
  updated_at: string;
};

export type PlanDistributionEntry = { tier: SubscriptionTier; count: number };
export type MonthlyRevenueEntry = { month: string; revenue_cents: number };
export type AdminStats = {
  total_active_subscriptions: number;
  total_revenue_cents: number;
  plan_distribution: PlanDistributionEntry[];
  monthly_revenue: MonthlyRevenueEntry[];
};
