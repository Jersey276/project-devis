export type QuoteStatus = "draft" | "sent" | "signed";

export type QuoteVat = {
  name: string;
  rate: number;
};

export type QuoteItem = {
  id: string;
  description: string;
  quantity: number;
  unitPrice: number;
  vat: QuoteVat;
};

export type Quote = {
  uuid: string;
  status: QuoteStatus;
  clientId: string;
  name: string;
  items: QuoteItem[];
};
