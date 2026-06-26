"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  createLine,
  deleteLine,
  type LineDraft,
  updateLine,
} from "@/lib/services/quotes";
import {
  createTemplate,
  createTemplateLine,
  deleteTemplate,
  listTemplateLines,
} from "@/lib/services/templates";
import type {
  BackendFee,
  BackendQuoteLine,
  BackendTax,
  BackendTemplateLine,
  QuoteLineData,
} from "@/types/backend";
import type {
  QuoteItemRow as RenderedRow,
  QuoteTotals,
} from "@/components/quote/steps/quote-step-items";

// ─── Types ───────────────────────────────────────────────────────────────────

export type FormItem = RenderedRow & { position: number };

// ─── Constants ───────────────────────────────────────────────────────────────

const SAVE_DEBOUNCE_MS = 600;
const SAVED_INDICATOR_MS = 1500;

// ─── Pure helpers ────────────────────────────────────────────────────────────

let _nextSublineKeyVal = 0;
function nextSublineKey(): string {
  return String(++_nextSublineKeyVal);
}

function normalizeLineData(
  data: QuoteLineData | undefined,
  lineType: BackendQuoteLine["type"],
): QuoteLineData {
  return {
    ...data,
    kind: data?.kind ?? (lineType === "multiple" ? "detailed" : "line"),
    sublines: data?.sublines?.map((s) =>
      s._key ? s : { ...s, _key: nextSublineKey() },
    ),
  };
}

function lineKind(item: FormItem): QuoteLineData["kind"] {
  return item.data.kind ?? (item.data.sublines?.length ? "detailed" : "line");
}

function leafAmount(item: FormItem): number {
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
}

export function computeTotals(
  items: FormItem[],
  taxById: Map<number, BackendTax>,
): QuoteTotals {
  const childrenByParent = new Map<string, FormItem[]>();
  for (const item of items) {
    const parentId = item.data.parent_line_id;
    if (!parentId) continue;
    const current = childrenByParent.get(parentId) ?? [];
    current.push(item);
    childrenByParent.set(parentId, current);
  }

  const visited = new Set<string>();

  const evalItem = (
    item: FormItem,
    taxIdOverride?: number | null,
  ): { principal: number; option: number; breakdown: Map<number, number> } => {
    if (visited.has(item.lineId)) {
      return { principal: 0, option: 0, breakdown: new Map() };
    }
    visited.add(item.lineId);

    const kind = lineKind(item);
    const taxId = taxIdOverride ?? item.taxId;
    const taxRate = taxId != null ? Number(taxById.get(taxId)?.rate ?? 0) : 0;
    const breakdown = new Map<number, number>();
    let principal = 0;
    let option = 0;

    if (kind === "detailed") {
      for (const subline of item.data.sublines ?? []) {
        const quantity = Number(subline.quantity);
        if (!Number.isFinite(quantity)) continue;
        const baseAmount = quantity * (subline.unit_price / 100);
        const taxAmount = baseAmount * (taxRate / 100);
        if (subline.option) {
          option += baseAmount;
        } else {
          principal += baseAmount;
          if (taxId != null) {
            breakdown.set(taxId, (breakdown.get(taxId) ?? 0) + taxAmount);
          }
        }
      }
    } else {
      const baseAmount = leafAmount(item);
      const taxAmount = baseAmount * (taxRate / 100);
      if (item.data.option) {
        option += baseAmount;
      } else {
        principal += baseAmount;
        if (taxId != null) {
          breakdown.set(taxId, (breakdown.get(taxId) ?? 0) + taxAmount);
        }
      }
    }

    for (const child of childrenByParent.get(item.lineId) ?? []) {
      const childTotals = evalItem(child);
      principal += childTotals.principal;
      option += childTotals.option;
      for (const [childTaxId, childAmount] of childTotals.breakdown.entries()) {
        breakdown.set(childTaxId, (breakdown.get(childTaxId) ?? 0) + childAmount);
      }
    }

    return { principal, option, breakdown };
  };

  const principalBreakdown = new Map<number, { tax: BackendTax; amount: number }>();
  let ht = 0;
  let optionHt = 0;

  for (const item of items) {
    if (item.data.parent_line_id) continue;
    const result = evalItem(item);
    ht += result.principal;
    optionHt += result.option;
    for (const [taxId, amount] of result.breakdown.entries()) {
      const tax = taxById.get(taxId);
      if (!tax) continue;
      const current = principalBreakdown.get(taxId);
      principalBreakdown.set(taxId, { tax, amount: (current?.amount ?? 0) + amount });
    }
  }

  const breakdown = Array.from(principalBreakdown.values()).sort(
    (a, b) => Number(a.tax.rate) - Number(b.tax.rate),
  );
  const ttc = ht + breakdown.reduce((acc, entry) => acc + entry.amount, 0);
  return { ht, breakdown, optionHt, optionTtc: optionHt, ttc };
}

// ─── Line/row converters ─────────────────────────────────────────────────────

export function rowFromBackendLine(line: BackendQuoteLine): FormItem {
  return {
    lineId: line.line_id,
    type: line.type,
    name: line.name,
    quantity: Number(line.quantity),
    unitPriceEuros: line.unit_price / 100,
    position: line.position,
    taxId: line.tax_id ?? null,
    data: normalizeLineData(line.data, line.type),
    saveStatus: "idle",
  };
}

function lineDraftFromRow(row: FormItem): LineDraft {
  const { data } = row;
  return {
    type: row.type,
    name: row.name,
    quantity: row.quantity,
    unitPriceEuros: row.unitPriceEuros,
    position: row.position,
    taxId: row.taxId,
    data: {
      ...data,
      sublines: data.sublines
        ?.filter((s) => s.name.trim() !== "" && s.quantity.trim() !== "")
        // eslint-disable-next-line @typescript-eslint/no-unused-vars
        .map(({ _key, ...rest }) => rest),
    },
  };
}

function buildLineDraft(
  kind: QuoteLineData["kind"],
  opts: { position: number; taxId: number | null; parentLineId?: string },
): LineDraft {
  const isDetailed = kind === "detailed";
  const isTextOrGroup = kind === "text" || kind === "group";
  return {
    type: isDetailed ? "multiple" : "simple",
    name: "",
    quantity: isTextOrGroup ? 0 : 1,
    unitPriceEuros: 0,
    position: opts.position,
    taxId: isTextOrGroup ? null : opts.taxId,
    data: {
      ...(kind !== "line" && { kind }),
      ...(opts.parentLineId && { parent_line_id: opts.parentLineId }),
      ...(isDetailed && { sublines: [] }),
      ...(kind === "text" && { description: "" }),
    },
  };
}

function newItemFromDraft(lineId: string, draft: LineDraft, position: number): FormItem {
  return {
    lineId,
    type: draft.type,
    name: "",
    quantity: draft.quantity,
    unitPriceEuros: 0,
    position,
    taxId: draft.taxId,
    data: draft.data ?? {},
    saveStatus: "idle",
  };
}

// ─── Hook ────────────────────────────────────────────────────────────────────

type UseQuoteLinesParams = {
  quoteId: string | undefined;
  isReadonly: boolean;
  defaultTaxId: number | null;
  items: FormItem[];
  setItems: React.Dispatch<React.SetStateAction<FormItem[]>>;
};

export function useQuoteLines({
  quoteId,
  isReadonly,
  defaultTaxId,
  items,
  setItems,
}: UseQuoteLinesParams) {
  const t = useTranslations("quote.form");

  const [adding, setAdding] = useState(false);

  const itemsRef = useRef(items);
  useEffect(() => { itemsRef.current = items; }, [items]);
  const defaultTaxIdRef = useRef(defaultTaxId);
  useEffect(() => { defaultTaxIdRef.current = defaultTaxId; }, [defaultTaxId]);

  const lineTimersRef = useRef(new Map<string, ReturnType<typeof setTimeout>>());
  const savedIndicatorTimersRef = useRef(new Map<string, ReturnType<typeof setTimeout>>());

  useEffect(() => {
    const lineTimers = lineTimersRef.current;
    const savedTimers = savedIndicatorTimersRef.current;
    return () => {
      for (const timer of lineTimers.values()) clearTimeout(timer);
      lineTimers.clear();
      for (const timer of savedTimers.values()) clearTimeout(timer);
      savedTimers.clear();
    };
  }, []);

  const setRow = useCallback((lineId: string, patch: Partial<FormItem>) => {
    setItems((prev) =>
      prev.map((row) => (row.lineId === lineId ? { ...row, ...patch } : row)),
    );
  }, [setItems]);

  const scheduleLineSave = useCallback(
    (lineId: string) => {
      if (!quoteId || isReadonly) return;
      const existingSaved = savedIndicatorTimersRef.current.get(lineId);
      if (existingSaved) {
        clearTimeout(existingSaved);
        savedIndicatorTimersRef.current.delete(lineId);
      }
      const existingTimer = lineTimersRef.current.get(lineId);
      if (existingTimer) clearTimeout(existingTimer);
      setRow(lineId, { saveStatus: "saving" });
      const timer = setTimeout(async () => {
        lineTimersRef.current.delete(lineId);
        const current = itemsRef.current.find((r) => r.lineId === lineId);
        if (!current) return;
        const { ok, body } = await updateLine(quoteId, lineId, lineDraftFromRow(current));
        if (ok && body.success) {
          setRow(lineId, { saveStatus: "saved" });
          const savedTimer = setTimeout(() => {
            savedIndicatorTimersRef.current.delete(lineId);
            setRow(lineId, { saveStatus: "idle" });
          }, SAVED_INDICATOR_MS);
          savedIndicatorTimersRef.current.set(lineId, savedTimer);
        } else {
          setRow(lineId, { saveStatus: "error" });
          toast.error((body.message as string) ?? t("errors.lineSaveFailedToast"));
        }
      }, SAVE_DEBOUNCE_MS);
      lineTimersRef.current.set(lineId, timer);
    },
    [isReadonly, quoteId, setRow, t],
  );

  const handleNameChange = useCallback(
    (lineId: string, value: string) => { setRow(lineId, { name: value }); scheduleLineSave(lineId); },
    [scheduleLineSave, setRow],
  );

  const handleQuantityChange = useCallback(
    (lineId: string, value: number) => { setRow(lineId, { quantity: Number.isFinite(value) ? value : 0 }); scheduleLineSave(lineId); },
    [scheduleLineSave, setRow],
  );

  const handleUnitPriceChange = useCallback(
    (lineId: string, value: number) => { setRow(lineId, { unitPriceEuros: Number.isFinite(value) ? value : 0 }); scheduleLineSave(lineId); },
    [scheduleLineSave, setRow],
  );

  const handleTaxChange = useCallback(
    (lineId: string, taxId: number | null) => { setRow(lineId, { taxId }); scheduleLineSave(lineId); },
    [scheduleLineSave, setRow],
  );

  const handleDescriptionChange = useCallback(
    (lineId: string, value: string) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId ? row : { ...row, data: { ...row.data, description: value } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleOptionChange = useCallback(
    (lineId: string, value: boolean) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId ? row : { ...row, data: { ...row.data, option: value } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleSublineAdd = useCallback(
    (lineId: string) => {
      const newSubline = { name: "", quantity: "1", unit_price: 0, option: false, _key: nextSublineKey() };
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : { ...row, data: { ...row.data, sublines: [...(row.data.sublines ?? []), newSubline] } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleSublineChange = useCallback(
    (lineId: string, index: number, patch: Partial<NonNullable<QuoteLineData["sublines"]>[number]>) => {
      setItems((prev) =>
        prev.map((row) => {
          if (row.lineId !== lineId) return row;
          const sublines = [...(row.data.sublines ?? [])];
          sublines[index] = { ...sublines[index], ...patch };
          return { ...row, data: { ...row.data, sublines } };
        }),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleSublineRemove = useCallback(
    (lineId: string, index: number) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : { ...row, data: { ...row.data, sublines: (row.data.sublines ?? []).filter((_, i) => i !== index) } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleAddItem = useCallback(
    async (kind: QuoteLineData["kind"] = "line") => {
      if (!quoteId || adding) return;
      setAdding(true);
      try {
        const draft = buildLineDraft(kind, { position: itemsRef.current.length, taxId: defaultTaxIdRef.current });
        const { ok, body } = await createLine(quoteId, draft);
        if (ok && body.success) {
          setItems((prev) => [...prev, newItemFromDraft(body.line_id as string, draft, prev.length)]);
        } else {
          toast.error((body.message as string) ?? t("errors.lineAddFailedToast"));
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, setItems, t],
  );

  const handleAddChildItem = useCallback(
    async (parentLineId: string, kind: QuoteLineData["kind"] = "line") => {
      if (!quoteId || adding) return;
      const parent = itemsRef.current.find((row) => row.lineId === parentLineId);
      if (!parent) return;
      setAdding(true);
      try {
        const draft = buildLineDraft(kind, { position: itemsRef.current.length, taxId: parent.taxId, parentLineId });
        const { ok, body } = await createLine(quoteId, draft);
        if (ok && body.success) {
          setItems((prev) => [...prev, newItemFromDraft(body.line_id as string, draft, prev.length)]);
        } else {
          toast.error((body.message as string) ?? t("errors.lineAddFailedToast"));
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, setItems, t],
  );

  const handleAddFeeItem = useCallback(
    async (fee: BackendFee) => {
      if (!quoteId || adding) return;
      setAdding(true);
      try {
        // A fee line snapshots the catalog entry; data.fee_id keeps the live
        // link so backend propagation can refresh it while the quote is a draft.
        const draft: LineDraft = {
          type: "simple",
          name: fee.name,
          quantity: 1,
          unit: fee.unit || undefined,
          unitPriceEuros: fee.unit_price / 100,
          position: itemsRef.current.length,
          taxId: fee.tax_id ?? null,
          data: { kind: "fee", fee_id: fee.fee_id },
        };
        const { ok, body } = await createLine(quoteId, draft);
        if (ok && body.success) {
          setItems((prev) => [
            ...prev,
            {
              lineId: body.line_id as string,
              type: "simple",
              name: fee.name,
              quantity: 1,
              unitPriceEuros: fee.unit_price / 100,
              position: prev.length,
              taxId: fee.tax_id ?? null,
              data: { kind: "fee", fee_id: fee.fee_id },
              saveStatus: "idle",
            },
          ]);
        } else {
          toast.error((body.message as string) ?? t("errors.lineAddFailedToast"));
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, setItems, t],
  );

  const handleAddFeeSubline = useCallback(
    (lineId: string, fee: BackendFee) => {
      const newSubline = {
        name: fee.name,
        quantity: "1",
        unit: fee.unit || undefined,
        unit_price: fee.unit_price,
        option: false,
        fee_id: fee.fee_id,
        _key: nextSublineKey(),
      };
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : { ...row, data: { ...row.data, sublines: [...(row.data.sublines ?? []), newSubline] } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setItems],
  );

  const handleRemoveItem = useCallback(
    async (lineId: string) => {
      if (!quoteId) return;
      const snapshot = itemsRef.current;
      const target = snapshot.find((r) => r.lineId === lineId);
      if (!target) return;
      if (
        !target.data.parent_line_id &&
        snapshot.filter((r) => !r.data.parent_line_id).length <= 1
      ) return;

      const toDelete = new Set<string>([lineId]);
      let frontier = [lineId];
      while (frontier.length > 0) {
        const next = snapshot
          .filter((r) => r.data.parent_line_id && frontier.includes(r.data.parent_line_id))
          .map((r) => r.lineId);
        next.forEach((id) => toDelete.add(id));
        frontier = next;
      }

      for (const id of toDelete) {
        const timer = lineTimersRef.current.get(id);
        if (timer) { clearTimeout(timer); lineTimersRef.current.delete(id); }
      }

      setItems((prev) => prev.filter((r) => !toDelete.has(r.lineId)));
      const results = await Promise.all([...toDelete].map((id) => deleteLine(quoteId, id)));
      if (results.some(({ ok, body }) => !ok || !body.success)) {
        toast.error(t("errors.lineRemoveFailedToast"));
        setItems(snapshot);
      }
    },
    [quoteId, setItems, t],
  );

  const handleAddItemFromTemplate = useCallback(
    async (templateId: string) => {
      if (!quoteId || adding) return;
      setAdding(true);
      try {
        const { ok, body } = await listTemplateLines(templateId);
        if (!ok || !Array.isArray(body.lines) || body.lines.length === 0) {
          toast.error(t("errors.lineAddFromTemplateFailedToast"));
          return;
        }
        const lines = (body.lines as BackendTemplateLine[]).sort((a, b) => a.position - b.position);
        const lineIdMap = new Map<string, string>();
        for (const tl of lines) {
          const draft: LineDraft = {
            type: tl.type === "multiple" ? "multiple" : "simple",
            name: tl.name,
            quantity: Number(tl.quantity),
            unit: tl.unit ?? undefined,
            unitPriceEuros: tl.unit_price / 100,
            position: itemsRef.current.length + lineIdMap.size,
            taxId: tl.tax_id ?? null,
            data: {
              ...tl.data,
              parent_line_id: tl.data.parent_line_id
                ? (lineIdMap.get(tl.data.parent_line_id) ?? tl.data.parent_line_id)
                : undefined,
            },
          };
          const createRes = await createLine(quoteId, draft);
          if (!createRes.ok || !createRes.body.success) {
            toast.error((createRes.body.message as string) ?? t("errors.lineAddFromTemplateFailedToast"));
            break;
          }
          const newLineId = createRes.body.line_id as string;
          lineIdMap.set(tl.line_id, newLineId);
          setItems((prev) => [
            ...prev,
            {
              lineId: newLineId,
              type: tl.type === "multiple" ? "multiple" : "simple",
              name: tl.name,
              quantity: Number(tl.quantity),
              unitPriceEuros: tl.unit_price / 100,
              position: prev.length,
              taxId: tl.tax_id ?? null,
              data: {
                ...draft.data,
                kind: draft.data?.kind ?? (tl.type === "multiple" ? "detailed" : "line"),
                sublines: draft.data?.sublines?.map((s) =>
                  s._key ? s : { ...s, _key: nextSublineKey() },
                ),
              },
              saveStatus: "idle",
            },
          ]);
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, setItems, t],
  );

  const handleSaveLineAsTemplate = useCallback(
    async (lineId: string, name: string): Promise<boolean> => {
      const row = itemsRef.current.find((r) => r.lineId === lineId);
      if (!row) return false;
      const { ok, body } = await createTemplate({ templateType: "quote_line", targetResource: "quote", name });
      if (!ok || !body.success) {
        toast.error((body.message as string) ?? t("errors.saveLineAsTemplateFailedToast"));
        return false;
      }
      const templateId = body.template_id as string;
      const lineRes = await createTemplateLine(templateId, {
        type: row.type, name: row.name, quantity: row.quantity,
        unitPriceEuros: row.unitPriceEuros, position: 0, taxId: row.taxId,
        data: { ...row.data, parent_line_id: undefined },
      });
      if (!lineRes.ok || !lineRes.body.success) {
        await deleteTemplate(templateId);
        toast.error((lineRes.body.message as string) ?? t("errors.saveLineAsTemplateFailedToast"));
        return false;
      }
      toast.success(t("saveAsTemplateSuccessToast"));
      return true;
    },
    [t],
  );

  const handleSaveQuoteAsTemplate = useCallback(
    async (name: string): Promise<boolean> => {
      const { ok, body } = await createTemplate({ templateType: "quote_document", targetResource: "quote", name });
      if (!ok || !body.success) {
        toast.error((body.message as string) ?? t("errors.saveQuoteAsTemplateFailedToast"));
        return false;
      }
      const templateId = body.template_id as string;
      const lineIdMap = new Map<string, string>();
      for (const [idx, row] of itemsRef.current.entries()) {
        const templateParentId = row.data.parent_line_id
          ? (lineIdMap.get(row.data.parent_line_id) ?? row.data.parent_line_id)
          : undefined;
        const lineRes = await createTemplateLine(templateId, {
          type: row.type, name: row.name, quantity: row.quantity,
          unitPriceEuros: row.unitPriceEuros, position: idx, taxId: row.taxId,
          data: { ...row.data, parent_line_id: templateParentId },
        });
        if (!lineRes.ok || !lineRes.body.success) {
          await deleteTemplate(templateId);
          toast.error((lineRes.body.message as string) ?? t("errors.saveQuoteAsTemplateFailedToast"));
          return false;
        }
        lineIdMap.set(row.lineId, lineRes.body.line_id as string);
      }
      toast.success(t("saveAsTemplateSuccessToast"));
      return true;
    },
    [t],
  );

  return {
    adding,
    setRow,
    scheduleLineSave,
    handleNameChange,
    handleQuantityChange,
    handleUnitPriceChange,
    handleTaxChange,
    handleDescriptionChange,
    handleOptionChange,
    handleSublineAdd,
    handleSublineChange,
    handleSublineRemove,
    handleAddItem,
    handleAddChildItem,
    handleAddFeeItem,
    handleAddFeeSubline,
    handleRemoveItem,
    handleAddItemFromTemplate,
    handleSaveLineAsTemplate,
    handleSaveQuoteAsTemplate,
  };
}
