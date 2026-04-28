export type QuoteFixture = {
  quote_id: string;
  user_id: string;
  name: string;
  archived_at: string | null;
  state: "draft" | "sent" | "validated" | "drop";
  created_at: string;
  updated_at: string;
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
};

export function quote(over: Partial<QuoteFixture> = {}): QuoteFixture {
  return {
    quote_id: "q-1",
    user_id: "u-1",
    name: "Devis Alpha",
    archived_at: null,
    state: "draft",
    created_at: "2026-01-01T00:00:00Z",
    updated_at: "2026-01-01T00:00:00Z",
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
    ...over,
  };
}
