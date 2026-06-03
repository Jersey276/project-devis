"use client";

import { Fragment, useMemo, useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
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
  ChevronDownIcon,
  ChevronRightIcon,
  EllipsisVerticalIcon,
  LayoutTemplateIcon,
  Loader2Icon,
  PlusIcon,
  Trash2Icon,
  TriangleAlertIcon,
} from "lucide-react";
import type {
  BackendTax,
  BackendQuoteLineType,
  QuoteLineData,
} from "@/types/backend";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import SelectLineTemplatePopover from "@/components/template/select-line-template-popover";

export type LineSaveStatus = "idle" | "saving" | "saved" | "error";

export type QuoteItemRow = {
  lineId: string;
  type: BackendQuoteLineType;
  name: string;
  quantity: number;
  unitPriceEuros: number;
  taxId: number | null;
  saveStatus: LineSaveStatus;
  data: QuoteLineData;
};

export type QuoteTotals = {
  ht: number;
  breakdown: Array<{ tax: BackendTax; amount: number }>;
  optionHt: number;
  optionTtc: number;
  ttc: number;
};

type QuoteStepItemsProps = {
  items: QuoteItemRow[];
  isReadonly: boolean;
  totals: QuoteTotals;
  availableTaxes: BackendTax[];
  taxById: Map<number, BackendTax>;
  isAdding: boolean;
  hideTotals?: boolean;
  onNameChange: (lineId: string, value: string) => void;
  onQuantityChange: (lineId: string, value: number) => void;
  onUnitPriceChange: (lineId: string, value: number) => void;
  onTaxChange: (lineId: string, taxId: number | null) => void;
  onDescriptionChange?: (lineId: string, value: string) => void;
  onOptionChange?: (lineId: string, value: boolean) => void;
  onRemoveItem: (lineId: string) => void;
  onAddItem: () => void;
  onSaveLineAsTemplate?: (lineId: string, name: string) => Promise<boolean>;
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
  hideTotals,
  onNameChange,
  onQuantityChange,
  onUnitPriceChange,
  onTaxChange,
  onDescriptionChange,
  onOptionChange,
  onRemoveItem,
  onAddItem,
  onSaveLineAsTemplate,
  onAddItemFromTemplate,
}: QuoteStepItemsProps) {
  const tCommon = useTranslations("common");
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
  const [expandedLineId, setExpandedLineId] = useState<string | null>(null);

  const depthByLineId = useMemo(() => {
    const parentById = new Map<string, string>();
    for (const item of items) {
      if (item.data.parent_line_id) {
        parentById.set(item.lineId, item.data.parent_line_id);
      }
    }
    const cache = new Map<string, number>();
    const depthOf = (lineId: string): number => {
      const cached = cache.get(lineId);
      if (cached !== undefined) return cached;
      const parentId = parentById.get(lineId);
      if (!parentId) {
        cache.set(lineId, 0);
        return 0;
      }
      const depth = depthOf(parentId) + 1;
      cache.set(lineId, depth);
      return depth;
    };
    for (const item of items) depthOf(item.lineId);
    return cache;
  }, [items]);

  const lineTotal = (item: QuoteItemRow): number => {
    const kind = lineKind(item);
    if (kind === "text" || kind === "group") return 0;
    if (kind === "detailed") {
      return (item.data.sublines ?? []).reduce((acc, subline) => {
        const quantity = Number(subline.quantity);
        if (!Number.isFinite(quantity)) return acc;
        return acc + quantity * (subline.unit_price / 100);
      }, 0);
    }
    return item.quantity * item.unitPriceEuros;
  };

  function lineKind(item: QuoteItemRow): QuoteLineData["kind"] {
    if (item.data.kind) return item.data.kind;
    if (item.data.sublines?.length) return "detailed";
    return "line";
  }

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
            const kind = lineKind(item);
            const depth = depthByLineId.get(item.lineId) ?? 0;
            const lineTotalValue = lineTotal(item);
            const selectedTax =
              item.taxId != null ? (taxById.get(item.taxId) ?? null) : null;
            const canEditAdvanced = !!onDescriptionChange && !!onOptionChange;
            const showAdvanced = expandedLineId === item.lineId;

            return (
              <Fragment key={item.lineId}>
                <TableRow
                  key={item.lineId}
                  data-line-id={item.lineId}
                  data-line-kind={kind}
                >
                  <TableCell>
                    <div
                      style={{ paddingLeft: depth * 16 }}
                      className="space-y-2"
                    >
                      <div className="flex items-center gap-2">
                        {depth > 0 && (
                          <span className="text-muted-foreground">↳</span>
                        )}
                        <Input
                          name="line-name"
                          value={item.name}
                          onChange={(event) =>
                            onNameChange(item.lineId, event.target.value)
                          }
                          disabled={isReadonly}
                          placeholder={t("lineNamePlaceholder")}
                        />
                        {kind !== "line" && (
                          <span className="text-muted-foreground rounded-full border px-2 py-0.5 text-xs">
                            {kind}
                          </span>
                        )}
                      </div>
                      {canEditAdvanced && (
                        <Button
                          type="button"
                          variant="ghost"
                          size="sm"
                          className="h-7 px-2 text-xs"
                          onClick={() =>
                            setExpandedLineId((current) =>
                              current === item.lineId ? null : item.lineId,
                            )
                          }
                        >
                          {showAdvanced ? (
                            <ChevronDownIcon className="mr-1 size-3.5" />
                          ) : (
                            <ChevronRightIcon className="mr-1 size-3.5" />
                          )}
                          {tCommon("actions.edit")}
                        </Button>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    {kind === "text" || kind === "group" ? (
                      <span className="text-muted-foreground">—</span>
                    ) : (
                      <Input
                        name="line-quantity"
                        type="number"
                        min={0}
                        value={item.quantity}
                        onChange={(event) =>
                          onQuantityChange(
                            item.lineId,
                            Number(event.target.value),
                          )
                        }
                        disabled={isReadonly || kind === "detailed"}
                      />
                    )}
                  </TableCell>
                  <TableCell>
                    {kind === "text" || kind === "group" ? (
                      <span className="text-muted-foreground">—</span>
                    ) : (
                      <Input
                        name="line-unit-price"
                        type="number"
                        min={0}
                        step="0.01"
                        value={item.unitPriceEuros}
                        onChange={(event) =>
                          onUnitPriceChange(
                            item.lineId,
                            Number(event.target.value),
                          )
                        }
                        disabled={isReadonly || kind === "detailed"}
                      />
                    )}
                  </TableCell>
                  <TableCell data-slot="line-tax-cell">
                    {kind === "text" || kind === "group" ? (
                      <span className="text-muted-foreground">—</span>
                    ) : (
                      <Combobox
                        items={availableTaxes}
                        value={selectedTax}
                        onValueChange={(t: BackendTax | null) =>
                          onTaxChange(item.lineId, t ? t.id : null)
                        }
                        itemToStringLabel={taxLabel}
                        disabled={
                          isReadonly || taxesDisabled || kind === "detailed"
                        }
                      >
                        <ComboboxInput
                          name="line-tax"
                          placeholder={
                            taxesDisabled ? "—" : t("taxPlaceholder")
                          }
                          disabled={
                            isReadonly || taxesDisabled || kind === "detailed"
                          }
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
                    )}
                  </TableCell>
                  <TableCell>{lineTotalValue.toFixed(2)} €</TableCell>
                  <TableCell>
                    <SaveIndicator
                      status={item.saveStatus}
                      labels={indicatorLabels}
                    />
                  </TableCell>
                  <TableCell>
                    {!isReadonly && (
                      <div className="flex items-center justify-end gap-1">
                        {onSaveLineAsTemplate && (
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
                              <DropdownMenuItem
                                onClick={() => openSaveLineDialog(item.lineId)}
                              >
                                <BookmarkIcon className="size-4" />
                                {t("saveAsTemplate")}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        )}
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          aria-label={t("deleteAria")}
                          disabled={items.length <= 1}
                          onClick={() => onRemoveItem(item.lineId)}
                        >
                          <Trash2Icon className="size-4" />
                        </Button>
                      </div>
                    )}
                  </TableCell>
                </TableRow>
                {showAdvanced && canEditAdvanced && (
                  <TableRow key={`${item.lineId}-advanced`}>
                    <TableCell colSpan={7} className="bg-muted/20">
                      <div className="grid gap-4 md:grid-cols-2">
                        <div className="space-y-1.5">
                          <Label htmlFor={`line-description-${item.lineId}`}>
                            Description
                          </Label>
                          <Textarea
                            id={`line-description-${item.lineId}`}
                            value={item.data.description ?? ""}
                            onChange={(event) =>
                              onDescriptionChange(
                                item.lineId,
                                event.target.value,
                              )
                            }
                            disabled={isReadonly || kind === "subline"}
                            placeholder="Description visible sur le devis"
                          />
                        </div>
                        <div className="flex items-end gap-3">
                          <div className="flex items-center gap-2 rounded-md border px-3 py-2">
                            <Checkbox
                              id={`line-option-${item.lineId}`}
                              checked={!!item.data.option}
                              onCheckedChange={(checked) =>
                                onOptionChange(item.lineId, checked === true)
                              }
                              disabled={
                                isReadonly ||
                                kind === "text" ||
                                kind === "group"
                              }
                            />
                            <Label htmlFor={`line-option-${item.lineId}`}>
                              Option
                            </Label>
                          </div>
                        </div>
                      </div>
                    </TableCell>
                  </TableRow>
                )}
              </Fragment>
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
        {!hideTotals && (
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
        )}
      </div>

      {onSaveLineAsTemplate && (
        <SaveTemplateDialog
          open={saveDialogOpen}
          onOpenChange={setSaveDialogOpen}
          defaultName={saveDialogDefaultName}
          onSave={async (name) => {
            if (saveDialogLineId) {
              return onSaveLineAsTemplate(saveDialogLineId, name);
            }
            return false;
          }}
        />
      )}
    </div>
  );
}
