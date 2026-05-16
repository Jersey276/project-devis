import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import {
  CheckIcon,
  Loader2Icon,
  PlusIcon,
  TriangleAlertIcon,
  Trash2Icon,
} from "lucide-react";
import type { BackendTax } from "@/types/backend";

export type LineSaveStatus = "idle" | "saving" | "saved" | "error";

export type QuoteItemRow = {
  lineId: string;
  name: string;
  quantity: number;
  unitPriceEuros: number;
  taxId: number | null;
  saveStatus: LineSaveStatus;
};

export type QuoteTotals = {
  ht: number;
  breakdown: Array<{ tax: BackendTax; amount: number }>;
  ttc: number;
};

type QuoteStepItemsProps = {
  items: QuoteItemRow[];
  isReadonly: boolean;
  totals: QuoteTotals;
  availableTaxes: BackendTax[];
  taxById: Map<number, BackendTax>;
  isAdding: boolean;
  onNameChange: (lineId: string, value: string) => void;
  onQuantityChange: (lineId: string, value: number) => void;
  onUnitPriceChange: (lineId: string, value: number) => void;
  onTaxChange: (lineId: string, taxId: number | null) => void;
  onRemoveItem: (lineId: string) => void;
  onAddItem: () => void;
};

function SaveIndicator({ status }: { status: LineSaveStatus }) {
  if (status === "saving") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="saving"
        className="text-muted-foreground inline-flex items-center"
        aria-label="Enregistrement en cours"
      >
        <Loader2Icon className="size-4 animate-spin" />
      </span>
    );
  }
  if (status === "saved") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="saved"
        className="inline-flex items-center text-emerald-600"
        aria-label="Enregistré"
      >
        <CheckIcon className="size-4" />
      </span>
    );
  }
  if (status === "error") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="error"
        className="text-destructive inline-flex items-center"
        aria-label="Échec d'enregistrement"
      >
        <TriangleAlertIcon className="size-4" />
      </span>
    );
  }
  return (
    <span
      data-slot="line-save-indicator"
      data-status="idle"
      className="sr-only"
    >
      idle
    </span>
  );
}

function taxLabel(t: BackendTax): string {
  const base = `${t.name} (${t.rate}%)`;
  return t.superseded_at ? `${base} — version archivée` : base;
}

export default function QuoteStepItems({
  items,
  isReadonly,
  totals,
  availableTaxes,
  taxById,
  isAdding,
  onNameChange,
  onQuantityChange,
  onUnitPriceChange,
  onTaxChange,
  onRemoveItem,
  onAddItem,
}: QuoteStepItemsProps) {
  const selectableTaxes = availableTaxes.filter((t) => !t.superseded_at);
  const taxesDisabled = selectableTaxes.length === 0;
  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Description</TableHead>
            <TableHead>Quantité</TableHead>
            <TableHead>Prix unitaire</TableHead>
            <TableHead>TVA</TableHead>
            <TableHead>Total ligne</TableHead>
            <TableHead className="w-12">État</TableHead>
            <TableHead className="w-24">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const lineTotal = item.quantity * item.unitPriceEuros;
            const selectedTax =
              item.taxId != null ? (taxById.get(item.taxId) ?? null) : null;

            return (
              <TableRow key={item.lineId} data-line-id={item.lineId}>
                <TableCell>
                  <Input
                    name="line-name"
                    value={item.name}
                    onChange={(event) =>
                      onNameChange(item.lineId, event.target.value)
                    }
                    disabled={isReadonly}
                    placeholder="Prestation"
                  />
                </TableCell>
                <TableCell>
                  <Input
                    name="line-quantity"
                    type="number"
                    min={0}
                    value={item.quantity}
                    onChange={(event) =>
                      onQuantityChange(item.lineId, Number(event.target.value))
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell>
                  <Input
                    name="line-unit-price"
                    type="number"
                    min={0}
                    step="0.01"
                    value={item.unitPriceEuros}
                    onChange={(event) =>
                      onUnitPriceChange(item.lineId, Number(event.target.value))
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell data-slot="line-tax-cell">
                  <Combobox
                    items={availableTaxes}
                    value={selectedTax}
                    onValueChange={(t: BackendTax | null) =>
                      onTaxChange(item.lineId, t ? t.id : null)
                    }
                    itemToStringLabel={taxLabel}
                    disabled={isReadonly || taxesDisabled}
                  >
                    <ComboboxInput
                      name="line-tax"
                      placeholder={taxesDisabled ? "—" : "Sélectionner"}
                      disabled={isReadonly || taxesDisabled}
                    />
                    <ComboboxContent>
                      <ComboboxEmpty>Aucune taxe disponible.</ComboboxEmpty>
                      <ComboboxList>
                        {(t: BackendTax) => (
                          <ComboboxItem
                            key={t.id}
                            value={t}
                            disabled={!!t.superseded_at}
                          >
                            {taxLabel(t)}
                          </ComboboxItem>
                        )}
                      </ComboboxList>
                    </ComboboxContent>
                  </Combobox>
                </TableCell>
                <TableCell>{lineTotal.toFixed(2)} €</TableCell>
                <TableCell>
                  <SaveIndicator status={item.saveStatus} />
                </TableCell>
                <TableCell>
                  {!isReadonly && (
                    <Button
                      type="button"
                      variant="ghost"
                      onClick={() => onRemoveItem(item.lineId)}
                      disabled={items.length <= 1}
                      aria-label="Supprimer la ligne"
                    >
                      <Trash2Icon className="size-4" />
                    </Button>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      {!isReadonly && (
        <Button
          type="button"
          variant="ghost"
          className="h-auto w-full p-0"
          onClick={onAddItem}
          disabled={isAdding}
          aria-label="Ajouter une ligne"
        >
          <Skeleton className="flex h-14 w-full items-center justify-center">
            {isAdding ? (
              <Loader2Icon className="size-7 animate-spin" />
            ) : (
              <PlusIcon className="size-7" />
            )}
          </Skeleton>
        </Button>
      )}

      <div className="flex justify-end">
        <div
          data-slot="quote-totals"
          className="min-w-[260px] space-y-1 rounded-md border px-4 py-2 text-sm"
        >
          <div className="flex justify-between font-medium">
            <span>Montant total HT</span>
            <span data-slot="total-ht">{totals.ht.toFixed(2)} €</span>
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
              <span>Total TTC</span>
              <span data-slot="total-ttc">{totals.ttc.toFixed(2)} €</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
