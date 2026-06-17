"use client";

import { Fragment, useCallback, useMemo, useState } from "react";
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
  BackendFee,
  BackendTax,
  BackendQuoteLineType,
  QuoteLineData,
} from "@/types/backend";
import { CoinsIcon } from "lucide-react";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import SelectLineTemplatePopover from "@/components/template/select-line-template-popover";
import SelectFeePopover from "@/components/fees/select-fee-popover";

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
  onAddItem: (kind?: QuoteLineData["kind"]) => void;
  onAddChildItem?: (parentLineId: string, kind?: QuoteLineData["kind"]) => void;
  onAddFeeItem?: (fee: BackendFee) => void | Promise<void>;
  onAddFeeSubline?: (lineId: string, fee: BackendFee) => void;
  onSublineAdd?: (lineId: string) => void;
  onSublineChange?: (
    lineId: string,
    index: number,
    patch: Partial<NonNullable<QuoteLineData["sublines"]>[number]>,
  ) => void;
  onSublineRemove?: (lineId: string, index: number) => void;
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

const CREATABLE_KINDS: Array<QuoteLineData["kind"]> = [
  "line",
  "text",
  "group",
  "detailed",
];

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
  onAddChildItem,
  onAddFeeItem,
  onAddFeeSubline,
  onSublineAdd,
  onSublineChange,
  onSublineRemove,
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

  const kindLabel = useCallback(
    (kind: QuoteLineData["kind"]): string => {
      switch (kind) {
        case "text":
          return t("lineKinds.text");
        case "group":
          return t("lineKinds.group");
        case "detailed":
          return t("lineKinds.detailed");
        case "subline":
          return t("lineKinds.subline");
        case "fee":
          return t("lineKinds.fee");
        default:
          return t("lineKinds.line");
      }
    },
    [t],
  );

  const childrenByParentId = useMemo(() => {
    const result = new Map<string, QuoteItemRow[]>();
    for (const item of items) {
      if (item.data.parent_line_id) {
        const arr = result.get(item.data.parent_line_id) ?? [];
        arr.push(item);
        result.set(item.data.parent_line_id, arr);
      }
    }
    return result;
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

  const topLevelCount = items.filter((i) => !i.data.parent_line_id).length;

  function renderItemRows(item: QuoteItemRow) {
    const kind = lineKind(item);
    const itemChildren = childrenByParentId.get(item.lineId) ?? [];
    const sublines = item.data.sublines ?? [];
    const lineTotalValue = lineTotal(item);
    const selectedTax =
      item.taxId != null ? (taxById.get(item.taxId) ?? null) : null;
    const canEditAdvanced = !!onDescriptionChange && !!onOptionChange;
    const showAdvanced = expandedLineId === item.lineId;

    return (
      <Fragment key={item.lineId}>
        <TableRow data-line-id={item.lineId} data-line-kind={kind}>
          <TableCell>
            <div className="space-y-2">
              <div className="flex items-center gap-2">
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
                  onQuantityChange(item.lineId, Number(event.target.value))
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
                  onUnitPriceChange(item.lineId, Number(event.target.value))
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
                onValueChange={(tax: BackendTax | null) =>
                  onTaxChange(item.lineId, tax ? tax.id : null)
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
                    {(tax: BackendTax) => (
                      <ComboboxItem
                        key={tax.id}
                        value={tax}
                        disabled={!!tax.superseded_at}
                      >
                        {taxLabel(tax)}
                      </ComboboxItem>
                    )}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
            )}
          </TableCell>
          <TableCell>{lineTotalValue.toFixed(2)} €</TableCell>
          <TableCell>
            <SaveIndicator status={item.saveStatus} labels={indicatorLabels} />
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
                  disabled={!item.data.parent_line_id && topLevelCount <= 1}
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
                      onDescriptionChange!(item.lineId, event.target.value)
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
                        onOptionChange!(item.lineId, checked === true)
                      }
                      disabled={
                        isReadonly || kind === "text" || kind === "group"
                      }
                    />
                    <Label htmlFor={`line-option-${item.lineId}`}>Option</Label>
                  </div>
                </div>
              </div>
            </TableCell>
          </TableRow>
        )}

        {kind === "group" && (!isReadonly || itemChildren.length > 0) && (
          <TableRow key={`${item.lineId}-group-children`}>
            <TableCell colSpan={7} className="p-0 pb-2 pl-8">
              <div className="overflow-hidden rounded-md border">
                {itemChildren.length > 0 && (
                  <Table>
                    <TableBody>
                      {itemChildren.map((child) => renderItemRows(child))}
                    </TableBody>
                  </Table>
                )}
                {!isReadonly && onAddChildItem && (
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button
                        type="button"
                        variant="ghost"
                        className="h-auto w-full p-0"
                        disabled={isAdding}
                        aria-label={t("addLineInGroupAria")}
                        title={t("addChildTypeAria")}
                      >
                        <Skeleton className="flex h-8 w-full items-center justify-center">
                          {isAdding ? (
                            <Loader2Icon className="size-4 animate-spin" />
                          ) : (
                            <PlusIcon className="size-4" />
                          )}
                        </Skeleton>
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="start">
                      {CREATABLE_KINDS.map((k) => (
                        <DropdownMenuItem
                          key={k}
                          onClick={() => onAddChildItem(item.lineId, k)}
                        >
                          {kindLabel(k)}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                )}
              </div>
            </TableCell>
          </TableRow>
        )}
        {kind === "detailed" && (!isReadonly || sublines.length > 0) && (
          <TableRow key={`${item.lineId}-sublines`}>
            <TableCell colSpan={7} className="p-0 pb-2 pl-8">
              <div className="overflow-hidden rounded-md border">
                {sublines.length > 0 && (
                  <Table>
                    <TableHeader>
                      <TableRow>
                        <TableHead>{t("sublines.name")}</TableHead>
                        <TableHead>{t("sublines.quantity")}</TableHead>
                        <TableHead>{t("sublines.unit")}</TableHead>
                        <TableHead>{t("sublines.unitPrice")}</TableHead>
                        <TableHead>{t("sublines.option")}</TableHead>
                        <TableHead className="w-10" />
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {sublines.map((subline, idx) => (
                        <TableRow key={subline._key ?? String(idx)}>
                          <TableCell>
                            <Input
                              value={subline.name}
                              onChange={(e) =>
                                onSublineChange?.(item.lineId, idx, {
                                  name: e.target.value,
                                })
                              }
                              disabled={isReadonly}
                              placeholder={t("lineNamePlaceholder")}
                            />
                          </TableCell>
                          <TableCell>
                            <Input
                              type="number"
                              min={0}
                              value={subline.quantity}
                              onChange={(e) =>
                                onSublineChange?.(item.lineId, idx, {
                                  quantity: e.target.value,
                                })
                              }
                              disabled={isReadonly}
                            />
                          </TableCell>
                          <TableCell>
                            <Input
                              value={subline.unit ?? ""}
                              onChange={(e) =>
                                onSublineChange?.(item.lineId, idx, {
                                  unit: e.target.value || undefined,
                                })
                              }
                              disabled={isReadonly}
                              placeholder={t("sublines.unitPlaceholder")}
                            />
                          </TableCell>
                          <TableCell>
                            <Input
                              type="number"
                              min={0}
                              step="0.01"
                              value={subline.unit_price / 100}
                              onChange={(e) =>
                                onSublineChange?.(item.lineId, idx, {
                                  unit_price: Math.round(
                                    Number(e.target.value) * 100,
                                  ),
                                })
                              }
                              disabled={isReadonly}
                            />
                          </TableCell>
                          <TableCell>
                            <Checkbox
                              checked={!!subline.option}
                              onCheckedChange={(checked) =>
                                onSublineChange?.(item.lineId, idx, {
                                  option: checked === true,
                                })
                              }
                              disabled={isReadonly}
                            />
                          </TableCell>
                          <TableCell>
                            {!isReadonly && (
                              <Button
                                type="button"
                                variant="ghost"
                                size="icon"
                                aria-label={t("sublines.deleteAria")}
                                onClick={() =>
                                  onSublineRemove?.(item.lineId, idx)
                                }
                              >
                                <Trash2Icon className="size-4" />
                              </Button>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                )}
                {!isReadonly && (onSublineAdd || onAddFeeSubline) && (
                  <div className="flex">
                    {onSublineAdd && (
                      <Button
                        type="button"
                        variant="ghost"
                        className="h-auto flex-1 p-0"
                        onClick={() => onSublineAdd(item.lineId)}
                        aria-label={t("sublines.addAria")}
                      >
                        <Skeleton className="flex h-8 w-full items-center justify-center">
                          <PlusIcon className="size-4" />
                        </Skeleton>
                      </Button>
                    )}
                    {onAddFeeSubline && (
                      <SelectFeePopover
                        onSelect={(fee) => onAddFeeSubline(item.lineId, fee)}
                      >
                        <Button
                          type="button"
                          variant="ghost"
                          className="h-auto p-0"
                          aria-label={t("sublines.addFeeAria")}
                        >
                          <Skeleton className="flex h-8 w-12 items-center justify-center">
                            <CoinsIcon className="size-4" />
                          </Skeleton>
                        </Button>
                      </SelectFeePopover>
                    )}
                  </div>
                )}
              </div>
            </TableCell>
          </TableRow>
        )}
      </Fragment>
    );
  }

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
          {items
            .filter((item) => !item.data.parent_line_id)
            .map((item) => renderItemRows(item))}
        </TableBody>
      </Table>

      {!isReadonly && (
        <div className="flex gap-2">
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                className="h-auto flex-1 p-0"
                disabled={isAdding}
                aria-label={t("addAria")}
                title={t("addTypeAria")}
              >
                <Skeleton className="flex h-14 w-full items-center justify-center">
                  {isAdding ? (
                    <Loader2Icon className="size-7 animate-spin" />
                  ) : (
                    <PlusIcon className="size-7" />
                  )}
                </Skeleton>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="start">
              {CREATABLE_KINDS.map((kind) => (
                <DropdownMenuItem key={kind} onClick={() => onAddItem(kind)}>
                  {kindLabel(kind)}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          {onAddFeeItem && (
            <SelectFeePopover
              disabled={isAdding}
              onSelect={(fee) => onAddFeeItem(fee)}
            >
              <Button
                type="button"
                variant="ghost"
                className="h-auto p-0"
                disabled={isAdding}
                aria-label={t("addFeeAria")}
              >
                <Skeleton className="flex h-14 w-14 items-center justify-center">
                  <CoinsIcon className="size-5" />
                </Skeleton>
              </Button>
            </SelectFeePopover>
          )}
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
