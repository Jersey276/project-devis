import type { Quote } from "@/types/backend";

export type QuoteMock = Quote;

export const quoteMocks: QuoteMock[] = [
  {
    uuid: "q-1001",
    status: "draft",
    name: "Refonte site",
    clientId: "",
    items: [
      {
        id: "it-1",
        description: "Design UI",
        quantity: 10,
        unitPrice: 80,
        vat: {
          name: "TVA 20%",
          rate: 20,
        },
      },
    ],
  },
  {
    uuid: "q-1002",
    status: "sent",
    name: "Application mobile",
    clientId: "",
    items: [
      {
        id: "it-2",
        description: "Développement",
        quantity: 30,
        unitPrice: 110,
        vat: {
          name: "TVA 20%",
          rate: 20,
        },
      },
    ],
  },
  {
    uuid: "q-1003",
    status: "signed",
    name: "Maintenance annuelle",
    clientId: "",
    items: [
      {
        id: "it-3",
        description: "Support mensuel",
        quantity: 12,
        unitPrice: 80,
        vat: {
          name: "TVA 20%",
          rate: 20,
        },
      },
    ],
  },
];

export function computeQuoteTotal(quote: Quote): number {
  return quote.items.reduce((acc, item) => {
    const lineBase = item.quantity * item.unitPrice;
    return acc + lineBase + (lineBase * item.vat.rate) / 100;
  }, 0);
}

export function getQuoteMockById(id: string): QuoteMock | undefined {
  return quoteMocks.find((quote) => quote.uuid === id);
}
