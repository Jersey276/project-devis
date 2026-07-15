export type QuoteFixture = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  state: "draft" | "sent" | "validated" | "drop";
  client_id: string;
  address_id: number;
  user_address_id: number;
  created_at: string;
  updated_at: string;
  total_ttc?: number;
};

export type LineFixture = {
  line_id: string;
  quote_id: string;
  type: "simple";
  name: string;
  quantity: string;
  unit: string;
  unit_price: number;
  data: Record<string, unknown>;
  position: number;
  tax_id: number | null;
};

export type TaxFixture = {
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

export function tax(over: Partial<TaxFixture> = {}): TaxFixture {
  return {
    id: 100,
    name: "TVA 20",
    rate: "20.00",
    country_group_id: 10,
    is_default: true,
    version: 1,
    ...over,
  };
}

export function quote(over: Partial<QuoteFixture> = {}): QuoteFixture {
  return {
    quote_id: "q-1",
    user_id: "u-1",
    name: "Devis Alpha",
    archived_at: null,
    state: "draft",
    // Empty by default: tests that exercise client/address pickers override
    // these explicitly. Keeping them empty means edit-mode tests don't fire an
    // unintended /me/clients/:id/addresses request.
    client_id: "",
    address_id: 0,
    user_address_id: 0,
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
    total_ttc: 0,
    ...over,
  };
}

export type ClientFixture = {
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
  client_type: "individual" | "business";
  archived: boolean;
  linked_user_id?: string;
};

export function client(over: Partial<ClientFixture> = {}): ClientFixture {
  return {
    client_id: "c-1",
    user_id: "u-1",
    first_name: "Jean",
    last_name: "Dupont",
    email: "jean@example.com",
    phone: "",
    company: "Acme",
    siren: "",
    siret: "",
    vat: "",
    client_type: "business",
    archived: false,
    ...over,
  };
}

export function line(over: Partial<LineFixture> = {}): LineFixture {
  return {
    line_id: "l-1",
    quote_id: "q-1",
    type: "simple",
    name: "Design UI",
    quantity: "10",
    unit: "",
    unit_price: 8000,
    data: {},
    position: 0,
    tax_id: null,
    ...over,
  };
}

export type InvoiceSummaryFixture = {
  invoice_id: string;
  invoice_number: string;
  status: "DRAFT" | "ISSUED" | "PAID" | "CANCELLED";
  quote_id: string;
  schedule_id: string;
  issued_at: string;
  due_date: string;
  total_ttc_cents: number;
  total_ht_cents: number;
  lifecycle_status: "NONE" | "DEPOSITED" | "RECEIVED" | "APPROVED" | "REJECTED" | "COLLECTED";
};

export function invoiceSummary(over: Partial<InvoiceSummaryFixture> = {}): InvoiceSummaryFixture {
  return {
    invoice_id: "inv-1",
    invoice_number: "2026-0001",
    status: "ISSUED",
    quote_id: "q-1",
    schedule_id: "",
    issued_at: "2026-01-15T10:00:00Z",
    due_date: "2026-02-14",
    total_ttc_cents: 12000,
    total_ht_cents: 10000,
    lifecycle_status: "NONE",
    ...over,
  };
}

export type InvoicePartyFixture = {
  company: string;
  first_name: string;
  last_name: string;
  siren: string;
  siret: string;
  vat: string;
  email: string;
  phone: string;
  street: string;
  additional_street: string;
  zip_code: string;
  city: string;
  iban: string;
  bic: string;
};

export function invoiceParty(over: Partial<InvoicePartyFixture> = {}): InvoicePartyFixture {
  return {
    company: "Acme SARL",
    first_name: "",
    last_name: "",
    siren: "123456782",
    siret: "12345678200010",
    vat: "FR12345678901",
    email: "acme@example.com",
    phone: "",
    street: "1 rue de Paris",
    additional_street: "",
    zip_code: "75001",
    city: "Paris",
    iban: "",
    bic: "",
    ...over,
  };
}

export type InvoiceLineFixture = {
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

export function invoiceLine(over: Partial<InvoiceLineFixture> = {}): InvoiceLineFixture {
  return {
    quote_line_id: "l-1",
    name: "Design UI",
    unit: "",
    quantity: "10",
    unit_price_cents: 8000,
    line_ht_cents: 10000,
    tax_id: 100,
    tax_rate: "20.00",
    tax_label: "TVA 20",
    ...over,
  };
}

export type InvoiceVatLineFixture = {
  tax_rate: string;
  base_ht_cents: number;
  vat_cents: number;
};

export function invoiceVatLine(over: Partial<InvoiceVatLineFixture> = {}): InvoiceVatLineFixture {
  return {
    tax_rate: "20.00",
    base_ht_cents: 10000,
    vat_cents: 2000,
    ...over,
  };
}

export type InvoiceDetailsFixture = {
  invoice_id: string;
  quote_id: string;
  schedule_id: string;
  billed_month_indexes: number[];
  status: "DRAFT" | "ISSUED" | "PAID" | "CANCELLED";
  invoice_number: string;
  issued_at: string;
  sale_date: string;
  due_date: string;
  issuer: InvoicePartyFixture;
  client: InvoicePartyFixture;
  lines: InvoiceLineFixture[];
  vat_breakdown: InvoiceVatLineFixture[];
  total_ht_cents: number;
  total_vat_cents: number;
  total_ttc_cents: number;
  vat_exempt: boolean;
  credited_positions: number[];
  lifecycle_status: "NONE" | "DEPOSITED" | "RECEIVED" | "APPROVED" | "REJECTED" | "COLLECTED";
};

export function invoiceDetails(over: Partial<InvoiceDetailsFixture> = {}): InvoiceDetailsFixture {
  return {
    invoice_id: "inv-1",
    quote_id: "q-1",
    schedule_id: "",
    billed_month_indexes: [],
    status: "ISSUED",
    invoice_number: "2026-0001",
    issued_at: "2026-01-15T10:00:00Z",
    sale_date: "2026-01-15",
    due_date: "2026-02-14",
    issuer: invoiceParty(),
    client: invoiceParty({
      company: "",
      first_name: "Jean",
      last_name: "Dupont",
      siren: "",
      siret: "",
      vat: "",
      email: "jean@example.com",
      street: "2 rue de Lyon",
      zip_code: "69001",
      city: "Lyon",
    }),
    lines: [invoiceLine()],
    vat_breakdown: [invoiceVatLine()],
    total_ht_cents: 10000,
    total_vat_cents: 2000,
    total_ttc_cents: 12000,
    vat_exempt: false,
    credited_positions: [],
    lifecycle_status: "NONE",
    ...over,
  };
}

export type CreditNoteSummaryFixture = {
  credit_note_id: string;
  credit_note_number: string;
  invoice_id: string;
  invoice_number: string;
  issued_at: string;
  is_total: boolean;
  total_ttc_cents: number;
};

export function creditNoteSummary(over: Partial<CreditNoteSummaryFixture> = {}): CreditNoteSummaryFixture {
  return {
    credit_note_id: "cn-1",
    credit_note_number: "A-2026-0001",
    invoice_id: "inv-1",
    invoice_number: "2026-0001",
    issued_at: "2026-01-20T10:00:00Z",
    is_total: false,
    total_ttc_cents: 12000,
    ...over,
  };
}

export type CreditNoteDetailsFixture = {
  credit_note_id: string;
  invoice_id: string;
  invoice_number: string;
  credit_note_number: string;
  issued_at: string;
  reason: string;
  is_total: boolean;
  issuer: InvoicePartyFixture;
  client: InvoicePartyFixture;
  lines: InvoiceLineFixture[];
  vat_breakdown: InvoiceVatLineFixture[];
  total_ht_cents: number;
  total_vat_cents: number;
  total_ttc_cents: number;
  vat_exempt: boolean;
};

export function creditNoteDetails(over: Partial<CreditNoteDetailsFixture> = {}): CreditNoteDetailsFixture {
  return {
    credit_note_id: "cn-1",
    invoice_id: "inv-1",
    invoice_number: "2026-0001",
    credit_note_number: "A-2026-0001",
    issued_at: "2026-01-20T10:00:00Z",
    reason: "Ligne annulée",
    is_total: false,
    issuer: invoiceParty(),
    client: invoiceParty({
      company: "",
      first_name: "Jean",
      last_name: "Dupont",
      siren: "",
      siret: "",
      vat: "",
      email: "jean@example.com",
      street: "2 rue de Lyon",
      zip_code: "69001",
      city: "Lyon",
    }),
    lines: [invoiceLine()],
    vat_breakdown: [invoiceVatLine()],
    total_ht_cents: 10000,
    total_vat_cents: 2000,
    total_ttc_cents: 12000,
    vat_exempt: false,
    ...over,
  };
}
