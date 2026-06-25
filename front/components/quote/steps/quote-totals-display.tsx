"use client";

import { useTranslations } from "next-intl";
import type { BackendTax } from "@/types/backend";
import type { QuoteTotals } from "./quote-step-items";

type Props = {
  totals: QuoteTotals;
  hideTotals?: boolean;
  taxLabel: (tax: BackendTax) => string;
};

export default function QuoteTotalsDisplay({
  totals,
  hideTotals,
  taxLabel,
}: Props) {
  const t = useTranslations("quote.steps.items");

  if (hideTotals) return <div className="flex justify-end" />;

  return (
    <div className="flex justify-end">
      <div
        data-slot="quote-totals"
        className="min-w-65 space-y-1 rounded-md border px-4 py-2 text-sm"
      >
        <div className="flex justify-between font-medium">
          <span>{t("totalHt")}</span>
          <span data-slot="total-ht">{totals.ht.toFixed(2)} €</span>
        </div>
        <div className="flex justify-between text-muted-foreground">
          <span>Total options</span>
          <span data-slot="total-option-ht">
            {totals.optionHt.toFixed(2)} €
          </span>
        </div>
        {totals.breakdown.map(({ tax, amount }) => (
          <div
            key={tax.id}
            data-slot="total-tax-line"
            data-tax-id={tax.id}
            className="text-muted-foreground flex justify-between"
          >
            <span>{taxLabel(tax)}</span>
            <span>{amount.toFixed(2)} €</span>
          </div>
        ))}
        {totals.breakdown.length > 0 && (
          <div className="flex justify-between border-t pt-1 font-semibold">
            <span>{t("totalTtc")}</span>
            <span data-slot="total-ttc">{totals.ttc.toFixed(2)} €</span>
          </div>
        )}
      </div>
    </div>
  );
}
