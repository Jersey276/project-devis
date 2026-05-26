"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  BookmarkIcon,
  CheckIcon,
  EllipsisVerticalIcon,
  LayoutTemplateIcon,
  Loader2Icon,
  PlusIcon,
  Trash2Icon,
  TriangleAlertIcon,
} from "lucide-react";
import type { BackendTax } from "@/types/backend";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import SelectLineTemplatePopover from "@/components/template/select-line-template-popover";

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
  onSaveLineAsTemplate?: (lineId: string, name: string) => Promise<void>;
  onAddItemFromTemplate?: (templateId: string) => Promise<void>;
};

type IndicatorLabels = { saving: string; saved: string; error: string };

function SaveIndicator({
  status,
  labels,
}: {
  status: LineSaveStatus;
  labels: IndicatorLabels;
}) {
  if (status === "saving") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="saving"
        className="text-muted-foreground inline-flex items-center"
        aria-label={labels.saving}
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
        aria-label={labels.saved}
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
        aria-label={labels.error}
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
  onSaveLineAsTemplate,
  onAddItemFromTemplate,
}: QuoteStepItemsProps) {
  const selectableTaxes = availableTaxes.filter((tax) => !tax.superseded_at);
  const taxesDisabled = selectableTaxes.length === 0;
  const t = useTranslations("quote.steps.items");
  const indicatorLabels: IndicatorLabels = {
    saving: t("savingAria"),
    saved: t("savedAria"),
    error: t("errorAria"),
  };

  const [saveDialogLineId, setSaveDialogLineId] = useState<string | null>(null);
  const [saveDialogOpen, setSaveDialogOpen] = useState(false);

  const taxLabel = (tax: BackendTax): string => {
    const base = `${tax.name} (${tax.rate}%)`;
    return tax.superseded_at ? `${base} — ${t("taxSupersededSuffix")}` : base;
  };

  function openSaveLineDialog(lineId: string) {
    setSaveDialogLineId(lineId);
    setSaveDialogOpen(true);
  }

  const saveDialogDefaultName =
    saveDialogLineId != null
      ? (items.find((i) => i.lineId === saveDialogLineId)?.name ?? "")
      : "";

  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("description")}</TableHead>
            <TableHead>{t("quantity")}</TableHead>
            <TableHead>{t("unitPrice")}</TableHead>
            <TableHead>{t("tax")}</TableHead>
            <TableHead>{t("lineTotal")}</TableHead>
            <TableHead className="w-12">{t("state")}</TableHead>
            <TableHead className="w-24">{t("action")}</TableHead>
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
                    placeholder={t("lineNamePlaceholder")}
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
                      placeholder={taxesDisabled ? "—" : t("taxPlaceholder")}
                      disabled={isReadonly || taxesDisabled}
                    />
                    <ComboboxContent>
                      <ComboboxEmpty>{t("taxEmpty")}</ComboboxEmpty>
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
                  <SaveIndicator
                    status={item.saveStatus}
                    labels={indicatorLabels}
                  />
                </TableCell>
                <TableCell>
                  {!isReadonly && (
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          aria-label={t("action")}
                        >
                          <EllipsisVerticalIcon className="size-4" />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        {onSaveLineAsTemplate && (
                          <DropdownMenuItem
                            onClick={() => openSaveLineDialog(item.lineId)}
                          >
                            <BookmarkIcon className="size-4" />
                            {t("saveAsTemplate")}
                          </DropdownMenuItem>
                        )}
                        <DropdownMenuItem
                          disabled={items.length <= 1}
                          className="text-destructive focus:text-destructive"
                          onClick={() => onRemoveItem(item.lineId)}
                        >
                          <Trash2Icon className="size-4" />
                          {t("deleteLine")}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      {!isReadonly && (
        <div className="flex gap-2">
          <Button
            type="button"
            variant="ghost"
            className="h-auto flex-1 p-0"
            onClick={onAddItem}
            disabled={isAdding}
            aria-label={t("addAria")}
          >
            <Skeleton className="flex h-14 w-full items-center justify-center">
              {isAdding ? (
                <Loader2Icon className="size-7 animate-spin" />
              ) : (
                <PlusIcon className="size-7" />
              )}
            </Skeleton>
          </Button>
          {onAddItemFromTemplate && (
            <SelectLineTemplatePopover
              disabled={isAdding}
              onSelect={onAddItemFromTemplate}
            >
              <Button
                type="button"
                variant="ghost"
                className="h-auto p-0"
                disabled={isAdding}
                aria-label={t("addFromTemplateAria")}
              >
                <Skeleton className="flex h-14 w-14 items-center justify-center">
                  <LayoutTemplateIcon className="size-5" />
                </Skeleton>
              </Button>
            </SelectLineTemplatePopover>
          )}
        </div>
      )}

      <div className="flex justify-end">
        <div
          data-slot="quote-totals"
          className="min-w-65 space-y-1 rounded-md border px-4 py-2 text-sm"
        >
          <div className="flex justify-between font-medium">
            <span>{t("totalHt")}</span>
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
              <span>{t("totalTtc")}</span>
              <span data-slot="total-ttc">{totals.ttc.toFixed(2)} €</span>
            </div>
          )}
        </div>
      </div>

      {onSaveLineAsTemplate && (
        <SaveTemplateDialog
          open={saveDialogOpen}
          onOpenChange={setSaveDialogOpen}
          defaultName={saveDialogDefaultName}
          onSave={async (name) => {
            if (saveDialogLineId) {
              await onSaveLineAsTemplate(saveDialogLineId, name);
            }
          }}
        />
      )}
    </div>
  );
}
