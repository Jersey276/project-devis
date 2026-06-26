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
