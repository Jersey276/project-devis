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
  ChevronDownIcon,
  ChevronRightIcon,
  EllipsisVerticalIcon,
  MessageSquareIcon,
  Trash2Icon,
} from "lucide-react";
import type {
  BackendFee,
  BackendTax,
  BackendQuoteLineType,
  QuoteLineData,
} from "@/types/backend";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import SaveIndicator, {
  type LineSaveStatus,
  type IndicatorLabels,
} from "./save-indicator";
import QuoteTotalsDisplay from "./quote-totals-display";
import LineDetailedSublines from "./line-detailed-sublines";
import LineGroupChildren from "./line-group-children";
import AddItemControls from "./add-item-controls";

export type { LineSaveStatus };

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
  onOpenComments?: (lineId: string, lineName: string) => void;
};

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
  onOpenComments,
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
    const lineTotalValue = lineTotal(item);
    const selectedTax =
      item.taxId != null ? (taxById.get(item.taxId) ?? null) : null;
    const canEditAdvanced = !!onDescriptionChange && !!onOptionChange;
    const showAdvanced = expandedLineId === item.lineId;
    // A fee line mirrors a catalog entry: its name/price/unit are driven by the
    // fee and may be overwritten by propagation, so they are locked here. The
    // tax stays editable — the fee only seeds a default tax — as do quantity
    // and the option flag.
    const isFee = kind === "fee";

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
                  disabled={isReadonly || isFee}
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
                disabled={isReadonly || kind === "detailed" || isFee}
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
            <div className="flex items-center justify-end gap-1">
              {onOpenComments && (
                <Button
                  type="button"
                  variant="ghost"
                  size="icon"
                  aria-label={t("commentsAria")}
                  onClick={() => onOpenComments(item.lineId, item.name)}
                >
                  <MessageSquareIcon className="size-4" />
                </Button>
              )}
              {!isReadonly && (
                <>
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
                </>
              )}
            </div>
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

        <LineGroupChildren
          item={item}
          isReadonly={isReadonly}
          isAdding={isAdding}
          renderedChildren={itemChildren.map((child) => renderItemRows(child))}
          kindLabel={kindLabel}
          onAddChildItem={onAddChildItem}
        />
        <LineDetailedSublines
          item={item}
          isReadonly={isReadonly}
          onSublineAdd={onSublineAdd}
          onSublineChange={onSublineChange}
          onSublineRemove={onSublineRemove}
          onAddFeeSubline={onAddFeeSubline}
        />
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
        <AddItemControls
          isAdding={isAdding}
          kindLabel={kindLabel}
          onAddItem={onAddItem}
          onAddFeeItem={onAddFeeItem}
          onAddItemFromTemplate={onAddItemFromTemplate}
        />
      )}

      <QuoteTotalsDisplay
        totals={totals}
        hideTotals={hideTotals}
        taxLabel={taxLabel}
      />

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
