"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import {
  DownloadIcon,
  BookmarkIcon,
  Loader2Icon,
  CalendarIcon,
} from "lucide-react";
import QuoteStepBasicInfo from "@/components/quote/steps/quote-step-basic-info";
import QuoteStepItems, {
  type QuoteItemRow as RenderedRow,
  type QuoteTotals,
} from "@/components/quote/steps/quote-step-items";
import QuoteStepSummary from "@/components/quote/steps/quote-step-summary";
import {
  continueQuote,
  createLine,
  createQuote,
  deleteLine,
  dropQuote,
  getQuote,
  type LineDraft,
  updateLine,
  updateQuote,
} from "@/lib/services/quotes";
import { exportQuotePdf } from "@/lib/services/export";
import {
  createTemplate,
  createTemplateLine,
  deleteTemplate,
  listTemplateLines,
} from "@/lib/services/templates";
import type { BackendTemplateLine } from "@/types/backend";
import { listClients } from "@/lib/services/clients";
import { listAddresses } from "@/lib/services/addresses";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import { apiFetch, fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { useMode } from "@/lib/mode-context";
import {
  type BackendAddress,
  type BackendClient,
  type BackendQuote,
  type BackendQuoteLine,
  type BackendQuoteState,
  type BackendTax,
  type QuoteLineData,
} from "@/types/backend";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import CreateScheduleDialog from "@/components/schedule/create-schedule-dialog";

type FormItem = RenderedRow & { position: number };

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

function lineTaxAmount(amount: number, taxRate: number): number {
  return amount * (1 + taxRate / 100);
}

function computeTotals(
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
        breakdown.set(
          childTaxId,
          (breakdown.get(childTaxId) ?? 0) + childAmount,
        );
      }
    }

    return { principal, option, breakdown };
  };

  const principalBreakdown = new Map<
    number,
    { tax: BackendTax; amount: number }
  >();
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
      principalBreakdown.set(taxId, {
        tax,
        amount: (current?.amount ?? 0) + amount,
      });
    }
  }

  const breakdown = Array.from(principalBreakdown.values()).sort(
    (a, b) => Number(a.tax.rate) - Number(b.tax.rate),
  );
  const ttc = ht + breakdown.reduce((acc, entry) => acc + entry.amount, 0);
  return { ht, breakdown, optionHt, optionTtc: optionHt, ttc };
}

type QuoteFormProps = {
  quoteId?: string;
};

const STEP_KEYS = ["basicInfo", "items"] as const;

const SAVE_DEBOUNCE_MS = 600;
const SAVED_INDICATOR_MS = 1500;

const STATE_BADGE_VARIANT: Record<
  BackendQuoteState,
  "default" | "secondary" | "destructive"
> = {
  draft: "secondary",
  sent: "default",
  validated: "default",
  drop: "destructive",
};

function rowFromBackendLine(line: BackendQuoteLine): FormItem {
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
        .map(({ _key: _, ...rest }) => rest),
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

function newItemFromDraft(
  lineId: string,
  draft: LineDraft,
  position: number,
): FormItem {
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

export default function QuoteForm({ quoteId }: QuoteFormProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isCustomer } = useMode();
  const t = useTranslations("quote.form");
  const tSteps = useTranslations("quote.steps");
  const tStatus = useTranslations("status.quote");
  const tCommon = useTranslations("common");
  const isCreate = !quoteId;

  const initialStep = useMemo(() => {
    const fromQuery = Number(searchParams.get("step"));
    if (Number.isFinite(fromQuery) && fromQuery >= 1 && fromQuery <= 3) {
      return fromQuery - 1;
    }
    return 0;
    // searchParams is read once on mount; intentionally exclude from deps
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const [step, setStep] = useState(initialStep);
  const [loading, setLoading] = useState(!isCreate);
  const [notFound, setNotFound] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [clientId, setClientId] = useState("");
  const [addressId, setAddressId] = useState<number | null>(null);
  const [userAddressId, setUserAddressId] = useState<number | null>(null);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [addresses, setAddresses] = useState<BackendAddress[]>([]);
  const [userAddresses, setUserAddresses] = useState<BackendAddress[]>([]);
  const [items, setItems] = useState<FormItem[]>([]);
  const [availableTaxes, setAvailableTaxes] = useState<BackendTax[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [creating, setCreating] = useState(false);
  const [adding, setAdding] = useState(false);
  const [quoteState, setQuoteState] = useState<BackendQuoteState>("draft");
  const [transitioning, setTransitioning] = useState(false);
  const [isExporting, setIsExporting] = useState(false);
  const [saveTemplateOpen, setSaveTemplateOpen] = useState(false);
  const [createScheduleOpen, setCreateScheduleOpen] = useState(false);
  const [userId, setUserId] = useState("");

  const templateIdFromQuery = useMemo(
    () => searchParams.get("template") ?? null,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const isReadonly =
    isCustomer || quoteState === "validated" || quoteState === "drop";

  const itemsRef = useRef(items);
  itemsRef.current = items;
  const projectNameRef = useRef(projectName);
  projectNameRef.current = projectName;

  const lineTimersRef = useRef(
    new Map<string, ReturnType<typeof setTimeout>>(),
  );
  const savedIndicatorTimersRef = useRef(
    new Map<string, ReturnType<typeof setTimeout>>(),
  );
  const nameTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Initial fetch (edit mode only)
  useEffect(() => {
    if (!quoteId) return;
    let cancelled = false;
    getQuote(quoteId).then(({ ok, body }) => {
      if (cancelled) return;
      if (!ok || !body.success) {
        setNotFound(true);
        toast.error((body.message as string) ?? t("notFoundToast"));
        return;
      }
      const fetchedQuote = body.quote as BackendQuote;
      const fetchedLines = (body.lines ?? []) as BackendQuoteLine[];
      setProjectName(fetchedQuote.name);
      setQuoteState(fetchedQuote.state ?? "draft");
      setClientId(fetchedQuote.client_id ?? "");
      setAddressId(fetchedQuote.address_id ?? null);
      setUserAddressId(fetchedQuote.user_address_id || null);
      const sorted = [...fetchedLines].sort((a, b) => a.position - b.position);
      setItems(sorted.map(rowFromBackendLine));
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [quoteId, t]);

  // Load clients (always — needed for create and to display in edit)
  useEffect(() => {
    let cancelled = false;
    listClients().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.clients)) {
        setClients(body.clients as BackendClient[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  // Load the current user's addresses. Step 1 picks the provider
  // (prestataire) address from this list; step 2 uses it to scope the
  // available taxes. The listAddresses owner needs the concrete user_id, so
  // we fetch /me first — the gateway resolves the authenticated user there.
  useEffect(() => {
    if (isCustomer) return;
    let cancelled = false;
    (async () => {
      const meRes = await apiFetch("/api/users/me");
      if (cancelled || !meRes.ok || !meRes.body.success || !meRes.body.user) {
        return;
      }
      const meId = (meRes.body.user as { user_id: string }).user_id;
      setUserId(meId);
      const { ok, body } = await listAddresses({
        type: "user",
        userId: meId,
      });
      if (cancelled) return;
      if (ok && Array.isArray(body.addresses)) {
        setUserAddresses(body.addresses as BackendAddress[]);
      } else {
        setUserAddresses([]);
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [isCustomer]);

  // Load taxes available for the current user (resolved from their first
  // address). Existing lines may reference superseded taxes (the tax was
  // edited or retired after the quote was created); we forward those ids
  // as include_ids so the combobox can still render the historical label.
  // Gated on `loading` so we don't fire twice in edit mode (once before
  // getQuote resolves, once after items populate).
  //
  // The orphan list is filtered against the *current* (non-superseded)
  // taxes already loaded — never against ALL availableTaxes. Filtering
  // against the full set would drop the orphan from include_ids on the
  // next render, triggering a re-fetch that returns only currents and
  // erases the orphan from availableTaxes. Cycle would oscillate.
  const currentTaxIds = useMemo(
    () =>
      new Set(availableTaxes.filter((t) => !t.superseded_at).map((t) => t.id)),
    [availableTaxes],
  );
  const includeTaxIds = useMemo(() => {
    const ids = items
      .map((i) => i.taxId)
      .filter((id): id is number => id != null && !currentTaxIds.has(id));
    return [...new Set(ids)].sort((a, b) => a - b);
  }, [items, currentTaxIds]);
  const includeTaxIdsKey = includeTaxIds.join(",");

  useEffect(() => {
    if (loading) return;
    let cancelled = false;
    listAvailableTaxesForUser(includeTaxIds, userAddressId ?? undefined).then(
      ({ ok, body }) => {
        if (cancelled) return;
        if (ok && Array.isArray(body.taxes)) {
          setAvailableTaxes(body.taxes as BackendTax[]);
        } else {
          setAvailableTaxes([]);
        }
      },
    );
    return () => {
      cancelled = true;
    };
    // includeTaxIds is reference-unstable; key it via the joined string.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [loading, includeTaxIdsKey, userAddressId]);

  // Load addresses for the selected client; reset address when client changes.
  useEffect(() => {
    if (!clientId) {
      setAddresses([]);
      return;
    }
    let cancelled = false;
    listAddresses({ type: "client", clientId }).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.addresses)) {
        setAddresses(body.addresses as BackendAddress[]);
      } else {
        setAddresses([]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [clientId]);

  useEffect(() => {
    if (notFound) router.replace("/quote");
  }, [notFound, router]);

  // Cleanup timers on unmount
  useEffect(() => {
    const lineTimers = lineTimersRef.current;
    const savedTimers = savedIndicatorTimersRef.current;
    return () => {
      for (const timer of lineTimers.values()) clearTimeout(timer);
      lineTimers.clear();
      for (const timer of savedTimers.values()) clearTimeout(timer);
      savedTimers.clear();
      if (nameTimerRef.current) clearTimeout(nameTimerRef.current);
    };
  }, []);

  const taxById = useMemo(
    () => new Map(availableTaxes.map((t) => [t.id, t])),
    [availableTaxes],
  );

  const defaultTaxId = useMemo(
    () => availableTaxes.find((t) => t.is_default)?.id ?? null,
    [availableTaxes],
  );

  const totals = useMemo(() => computeTotals(items, taxById), [items, taxById]);

  // ────────────────────────────────────────────────────────────
  // Step 1 handlers

  const handleProjectNameChange = useCallback(
    (value: string) => {
      setProjectName(value);
      setErrors((prev) => ({ ...prev, name: [] }));
      if (!quoteId || isReadonly) return;
      if (nameTimerRef.current) clearTimeout(nameTimerRef.current);
      nameTimerRef.current = setTimeout(async () => {
        nameTimerRef.current = null;
        const current = projectNameRef.current.trim();
        if (current.length === 0) return;
        const { ok, body } = await updateQuote(quoteId, { name: current });
        if (!ok || !body.success) {
          toast.error(
            (body.message as string) ?? t("errors.nameSaveFailedToast"),
          );
        }
      }, SAVE_DEBOUNCE_MS);
    },
    [isReadonly, quoteId, t],
  );

  const handleNextFromStep1 = useCallback(async () => {
    const trimmed = projectName.trim();
    const localErrors: FieldErrors = {};
    if (trimmed.length === 0) {
      localErrors.name = [t("errors.requiredField")];
    }
    if (!clientId) {
      localErrors.client_id = [t("errors.selectClient")];
    }
    if (!addressId) {
      localErrors.address_id = [t("errors.selectAddress")];
    }
    if (!userAddressId) {
      localErrors.user_address_id = [t("errors.selectUserAddress")];
    }
    if (Object.keys(localErrors).length > 0) {
      setErrors(localErrors);
      return;
    }

    if (isCreate) {
      setCreating(true);
      try {
        const { ok, status, body } = await createQuote({
          name: trimmed,
          clientId,
          addressId: addressId!,
          userAddressId: userAddressId!,
        });
        if (ok && body.success) {
          const newId = body.quote_id as string;
          if (templateIdFromQuery) {
            const linesRes = await listTemplateLines(templateIdFromQuery);
            if (linesRes.ok && Array.isArray(linesRes.body.lines)) {
              const sorted = (
                linesRes.body.lines as BackendTemplateLine[]
              ).sort((a, b) => a.position - b.position);
              await Promise.all(
                sorted.map((tl, idx) =>
                  createLine(newId, {
                    type: tl.type === "multiple" ? "multiple" : "simple",
                    name: tl.name,
                    quantity: Number(tl.quantity),
                    unit: tl.unit ?? undefined,
                    unitPriceEuros: tl.unit_price / 100,
                    position: idx,
                    taxId: tl.tax_id ?? null,
                    data: tl.data,
                  }),
                ),
              );
            }
          }
          router.replace(`/quote/${newId}?step=2`);
          return;
        }
        if (status === 422 && Array.isArray(body.field_errors)) {
          setErrors(fieldErrorsFromBody(body));
        } else {
          toast.error((body.message as string) ?? tCommon("errors.generic"));
        }
      } catch {
        toast.error(tCommon("errors.generic"));
      } finally {
        setCreating(false);
      }
      return;
    }
    setStep(1);
  }, [
    addressId,
    clientId,
    isCreate,
    projectName,
    router,
    t,
    tCommon,
    templateIdFromQuery,
    userAddressId,
  ]);

  // ────────────────────────────────────────────────────────────
  // Step 2 handlers

  const setRow = useCallback((lineId: string, patch: Partial<FormItem>) => {
    setItems((prev) =>
      prev.map((row) => (row.lineId === lineId ? { ...row, ...patch } : row)),
    );
  }, []);

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
        const { ok, body } = await updateLine(
          quoteId,
          lineId,
          lineDraftFromRow(current),
        );
        if (ok && body.success) {
          setRow(lineId, { saveStatus: "saved" });
          const savedTimer = setTimeout(() => {
            savedIndicatorTimersRef.current.delete(lineId);
            setRow(lineId, { saveStatus: "idle" });
          }, SAVED_INDICATOR_MS);
          savedIndicatorTimersRef.current.set(lineId, savedTimer);
        } else {
          setRow(lineId, { saveStatus: "error" });
          toast.error(
            (body.message as string) ?? t("errors.lineSaveFailedToast"),
          );
        }
      }, SAVE_DEBOUNCE_MS);

      lineTimersRef.current.set(lineId, timer);
    },
    [isReadonly, quoteId, setRow, t],
  );

  const handleDrop = useCallback(async () => {
    if (!quoteId || transitioning) return;
    setTransitioning(true);
    try {
      const { ok, body } = await dropQuote(quoteId);
      if (ok && body.success) {
        setQuoteState("drop");
      } else {
        toast.error((body.message as string) ?? t("dropFailedToast"));
      }
    } finally {
      setTransitioning(false);
    }
  }, [quoteId, transitioning, t]);

  const handleContinue = useCallback(async () => {
    if (!quoteId || transitioning) return;
    setTransitioning(true);
    try {
      const { ok, body } = await continueQuote(quoteId);
      if (ok && body.success) {
        setQuoteState("draft");
      } else {
        toast.error((body.message as string) ?? t("continueFailedToast"));
      }
    } finally {
      setTransitioning(false);
    }
  }, [quoteId, transitioning, t]);

  const handleExport = useCallback(async () => {
    if (!quoteId || isExporting) return;
    setIsExporting(true);
    try {
      await exportQuotePdf(quoteId);
    } catch (err) {
      console.error("export quote pdf failed", err);
      toast.error(t("exportFailedToast"));
    } finally {
      setIsExporting(false);
    }
  }, [quoteId, isExporting, t]);

  const handleSaveQuoteAsTemplate = useCallback(
    async (name: string): Promise<boolean> => {
      const { ok, body } = await createTemplate({
        templateType: "quote_document",
        targetResource: "quote",
        name,
      });
      if (!ok || !body.success) {
        toast.error(
          (body.message as string) ??
            t("errors.saveQuoteAsTemplateFailedToast"),
        );
        return false;
      }
      const templateId = body.template_id as string;
      const lineIdMap = new Map<string, string>();
      for (const [idx, row] of itemsRef.current.entries()) {
        const templateParentId = row.data.parent_line_id
          ? (lineIdMap.get(row.data.parent_line_id) ?? row.data.parent_line_id)
          : undefined;
        const lineRes = await createTemplateLine(templateId, {
          type: row.type,
          name: row.name,
          quantity: row.quantity,
          unitPriceEuros: row.unitPriceEuros,
          position: idx,
          taxId: row.taxId,
          data: { ...row.data, parent_line_id: templateParentId },
        });
        if (!lineRes.ok || !lineRes.body.success) {
          await deleteTemplate(templateId);
          toast.error(
            (lineRes.body.message as string) ??
              t("errors.saveQuoteAsTemplateFailedToast"),
          );
          return false;
        }
        lineIdMap.set(row.lineId, lineRes.body.line_id as string);
      }
      toast.success(t("saveAsTemplateSuccessToast"));
      return true;
    },
    [t],
  );

  const handleSaveLineAsTemplate = useCallback(
    async (lineId: string, name: string): Promise<boolean> => {
      const row = itemsRef.current.find((r) => r.lineId === lineId);
      if (!row) return false;
      const { ok, body } = await createTemplate({
        templateType: "quote_line",
        targetResource: "quote",
        name,
      });
      if (!ok || !body.success) {
        toast.error(
          (body.message as string) ?? t("errors.saveLineAsTemplateFailedToast"),
        );
        return false;
      }
      const templateId = body.template_id as string;
      const lineRes = await createTemplateLine(templateId, {
        type: row.type,
        name: row.name,
        quantity: row.quantity,
        unitPriceEuros: row.unitPriceEuros,
        position: 0,
        taxId: row.taxId,
        data: { ...row.data, parent_line_id: undefined },
      });
      if (!lineRes.ok || !lineRes.body.success) {
        await deleteTemplate(templateId);
        toast.error(
          (lineRes.body.message as string) ??
            t("errors.saveLineAsTemplateFailedToast"),
        );
        return false;
      }
      toast.success(t("saveAsTemplateSuccessToast"));
      return true;
    },
    [t],
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
        const lines = (body.lines as BackendTemplateLine[]).sort(
          (a, b) => a.position - b.position,
        );
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
                ? (lineIdMap.get(tl.data.parent_line_id) ??
                  tl.data.parent_line_id)
                : undefined,
            },
          };
          const createRes = await createLine(quoteId, draft);
          if (!createRes.ok || !createRes.body.success) {
            toast.error(
              (createRes.body.message as string) ??
                t("errors.lineAddFromTemplateFailedToast"),
            );
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
                kind:
                  draft.data?.kind ??
                  (tl.type === "multiple" ? "detailed" : "line"),
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
    [adding, quoteId, t],
  );

  const handleAddItem = useCallback(
    async (kind: QuoteLineData["kind"] = "line") => {
      if (!quoteId || adding) return;
      setAdding(true);
      try {
        const draft = buildLineDraft(kind, {
          position: itemsRef.current.length,
          taxId: defaultTaxId,
        });
        const { ok, body } = await createLine(quoteId, draft);
        if (ok && body.success) {
          const newLineId = body.line_id as string;
          setItems((prev) => [
            ...prev,
            newItemFromDraft(newLineId, draft, prev.length),
          ]);
        } else {
          toast.error(
            (body.message as string) ?? t("errors.lineAddFailedToast"),
          );
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, defaultTaxId, t],
  );

  const handleAddChildItem = useCallback(
    async (parentLineId: string, kind: QuoteLineData["kind"] = "line") => {
      if (!quoteId || adding) return;
      const parent = itemsRef.current.find(
        (row) => row.lineId === parentLineId,
      );
      if (!parent) return;
      setAdding(true);
      try {
        const draft = buildLineDraft(kind, {
          position: itemsRef.current.length,
          taxId: parent.taxId,
          parentLineId,
        });
        const { ok, body } = await createLine(quoteId, draft);
        if (ok && body.success) {
          const newLineId = body.line_id as string;
          setItems((prev) => [
            ...prev,
            newItemFromDraft(newLineId, draft, prev.length),
          ]);
        } else {
          toast.error(
            (body.message as string) ?? t("errors.lineAddFailedToast"),
          );
        }
      } finally {
        setAdding(false);
      }
    },
    [adding, quoteId, t],
  );

  const handleRemoveItem = useCallback(
    async (lineId: string) => {
      if (!quoteId) return;
      const snapshot = itemsRef.current;
      const target = snapshot.find((r) => r.lineId === lineId);
      if (!target) return;
      // Block only when deleting the last top-level line.
      if (
        !target.data.parent_line_id &&
        snapshot.filter((r) => !r.data.parent_line_id).length <= 1
      )
        return;

      // BFS to collect the line and all its descendants.
      const toDelete = new Set<string>([lineId]);
      let frontier = [lineId];
      while (frontier.length > 0) {
        const next = snapshot
          .filter(
            (r) =>
              r.data.parent_line_id && frontier.includes(r.data.parent_line_id),
          )
          .map((r) => r.lineId);
        next.forEach((id) => toDelete.add(id));
        frontier = next;
      }

      // Cancel pending save timers.
      for (const id of toDelete) {
        const timer = lineTimersRef.current.get(id);
        if (timer) {
          clearTimeout(timer);
          lineTimersRef.current.delete(id);
        }
      }

      setItems((prev) => prev.filter((r) => !toDelete.has(r.lineId)));
      const results = await Promise.all(
        [...toDelete].map((id) => deleteLine(quoteId, id)),
      );
      if (results.some(({ ok, body }) => !ok || !body.success)) {
        toast.error(t("errors.lineRemoveFailedToast"));
        setItems(snapshot);
      }
    },
    [quoteId, t],
  );

  const handleNameChange = useCallback(
    (lineId: string, value: string) => {
      setRow(lineId, { name: value });
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setRow],
  );

  const handleQuantityChange = useCallback(
    (lineId: string, value: number) => {
      setRow(lineId, { quantity: Number.isFinite(value) ? value : 0 });
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setRow],
  );

  const handleUnitPriceChange = useCallback(
    (lineId: string, value: number) => {
      setRow(lineId, {
        unitPriceEuros: Number.isFinite(value) ? value : 0,
      });
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setRow],
  );

  const handleTaxChange = useCallback(
    (lineId: string, taxId: number | null) => {
      setRow(lineId, { taxId });
      scheduleLineSave(lineId);
    },
    [scheduleLineSave, setRow],
  );

  const handleDescriptionChange = useCallback(
    (lineId: string, value: string) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : { ...row, data: { ...row.data, description: value } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave],
  );

  const handleOptionChange = useCallback(
    (lineId: string, value: boolean) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : { ...row, data: { ...row.data, option: value } },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave],
  );

  const handleSublineAdd = useCallback(
    (lineId: string) => {
      const newSubline = {
        name: "",
        quantity: "1",
        unit_price: 0,
        option: false,
        _key: nextSublineKey(),
      };
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : {
                ...row,
                data: {
                  ...row.data,
                  sublines: [...(row.data.sublines ?? []), newSubline],
                },
              },
        ),
      );
      // The empty subline is stripped by lineDraftFromRow until the user types a name;
      // this save persists the line's other fields and clears any stale save indicator.
      scheduleLineSave(lineId);
    },
    [scheduleLineSave],
  );

  const handleSublineChange = useCallback(
    (
      lineId: string,
      index: number,
      patch: Partial<NonNullable<QuoteLineData["sublines"]>[number]>,
    ) => {
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
    [scheduleLineSave],
  );

  const handleSublineRemove = useCallback(
    (lineId: string, index: number) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId !== lineId
            ? row
            : {
                ...row,
                data: {
                  ...row.data,
                  sublines: (row.data.sublines ?? []).filter(
                    (_, i) => i !== index,
                  ),
                },
              },
        ),
      );
      scheduleLineSave(lineId);
    },
    [scheduleLineSave],
  );

  const handleClientIdChange = useCallback(
    (value: string) => {
      setClientId(value);
      setAddressId(null);
      setErrors((prev) => ({ ...prev, client_id: [], address_id: [] }));
      // Edit mode: persist immediately. Create mode persists on Suivant.
      // We can't write address_id=0 here because the user must pick one for
      // the new client; the next picker change will persist both.
      if (!quoteId || isReadonly || !value) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, clientId: value }).then(
        ({ ok, body }) => {
          if (!ok || !body.success) {
            toast.error(
              (body.message as string) ?? t("errors.clientSaveFailedToast"),
            );
          }
        },
      );
    },
    [isReadonly, quoteId, t],
  );

  const handleAddressIdChange = useCallback(
    (value: number | null) => {
      setAddressId(value);
      setErrors((prev) => ({ ...prev, address_id: [] }));
      if (!quoteId || isReadonly || value == null) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, addressId: value }).then(
        ({ ok, body }) => {
          if (!ok || !body.success) {
            toast.error(
              (body.message as string) ?? t("errors.addressSaveFailedToast"),
            );
          }
        },
      );
    },
    [isReadonly, quoteId, t],
  );

  const handleUserAddressIdChange = useCallback(
    (value: number | null) => {
      setUserAddressId(value);
      setErrors((prev) => ({ ...prev, user_address_id: [] }));
      if (!quoteId || isReadonly || value == null) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, userAddressId: value }).then(
        ({ ok, body }) => {
          if (!ok || !body.success) {
            toast.error(
              (body.message as string) ??
                t("errors.userAddressSaveFailedToast"),
            );
          }
        },
      );
    },
    [isReadonly, quoteId, t],
  );

  const refreshClients = useCallback(async () => {
    const { ok, body } = await listClients();
    if (ok && Array.isArray(body.clients)) {
      setClients(body.clients as BackendClient[]);
    }
  }, []);

  const refreshUserAddresses = useCallback(async () => {
    if (!userId) return;
    const { ok, body } = await listAddresses({ type: "user", userId });
    if (ok && Array.isArray(body.addresses)) {
      setUserAddresses(body.addresses as BackendAddress[]);
    }
  }, [userId]);

  const refreshClientAddresses = useCallback(async () => {
    if (!clientId) return;
    const { ok, body } = await listAddresses({ type: "client", clientId });
    if (ok && Array.isArray(body.addresses)) {
      setAddresses(body.addresses as BackendAddress[]);
    }
  }, [clientId]);

  // ────────────────────────────────────────────────────────────
  // Render

  if (loading) {
    return (
      <Card>
        <CardContent className="flex items-center justify-center py-16">
          <Loader2Icon
            data-slot="quote-form-loader"
            className="text-muted-foreground size-6 animate-spin"
          />
        </CardContent>
      </Card>
    );
  }

  const canGoNextFromStep1 =
    projectName.trim().length > 0 &&
    !!clientId &&
    !!addressId &&
    !!userAddressId &&
    !creating;

  const showDropButton =
    !isCustomer &&
    !isCreate &&
    (quoteState === "draft" || quoteState === "sent");
  const showContinueButton = !isCustomer && !isCreate && quoteState === "drop";

  return (
    <Card data-quote-state={quoteState}>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div className="space-y-1.5">
          <CardTitle className="flex items-center gap-2">
            {isCreate
              ? t("createTitle")
              : t("editTitlePlaceholder", { name: projectName || "…" })}
            {!isCreate && (
              <Badge
                data-slot="quote-state-badge"
                variant={STATE_BADGE_VARIANT[quoteState]}
              >
                {tStatus(quoteState)}
              </Badge>
            )}
          </CardTitle>
          <CardDescription>
            {isCreate ? t("createDescription") : t("editDescription")}
          </CardDescription>
        </div>
        {!isCreate && (
          <Button
            type="button"
            variant="outline"
            disabled={isExporting}
            onClick={handleExport}
          >
            <DownloadIcon className="size-4" />
            {t("exportButton")}
          </Button>
        )}
        {!isCreate && !isReadonly && (
          <Button
            type="button"
            variant="outline"
            onClick={() => setSaveTemplateOpen(true)}
          >
            <BookmarkIcon className="size-4" />
            {t("saveAsTemplateButton")}
          </Button>
        )}
        {!isCreate && !isCustomer && (
          <Button
            type="button"
            variant="outline"
            onClick={() => setCreateScheduleOpen(true)}
          >
            <CalendarIcon className="size-4" />
            Créer un échéancier
          </Button>
        )}
        {showDropButton && (
          <AlertDialog>
            <AlertDialogTrigger asChild>
              <Button
                type="button"
                variant="destructive"
                disabled={transitioning}
              >
                {t("dropButton")}
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>{t("dropDialog.title")}</AlertDialogTitle>
                <AlertDialogDescription>
                  {t("dropDialog.description")}
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>
                  {tCommon("actions.cancel")}
                </AlertDialogCancel>
                <AlertDialogAction variant="destructive" onClick={handleDrop}>
                  {t("dropDialog.confirm")}
                </AlertDialogAction>
              </AlertDialogFooter>
            </AlertDialogContent>
          </AlertDialog>
        )}
        {showContinueButton && (
          <Button
            type="button"
            onClick={handleContinue}
            disabled={transitioning}
          >
            {t("continueButton")}
          </Button>
        )}
      </CardHeader>

      <CardContent className="space-y-6">
        <div className="grid gap-2 sm:grid-cols-3">
          {STEP_KEYS.map((key, index) => (
            <Button
              key={key}
              type="button"
              variant="ghost"
              data-step-tab={index}
              data-active={step === index ? "true" : undefined}
              className={`justify-start rounded-md border px-3 py-2 text-sm font-normal ${
                step === index ? "bg-muted" : ""
              }`}
              onClick={() => {
                if (!isCreate) setStep(index);
              }}
              disabled={isCreate}
            >
              {t("stepLabel", { n: index + 1, label: tSteps(`${key}.label`) })}
            </Button>
          ))}
        </div>

        {step === 0 && (
          <QuoteStepBasicInfo
            projectName={projectName}
            clientId={clientId}
            addressId={addressId}
            userAddressId={userAddressId}
            isReadonly={isReadonly}
            clients={clients}
            addresses={addresses}
            userAddresses={userAddresses}
            userId={userId}
            nameErrors={errors.name}
            clientErrors={errors.client_id}
            addressErrors={errors.address_id}
            userAddressErrors={errors.user_address_id}
            onProjectNameChange={handleProjectNameChange}
            onClientIdChange={handleClientIdChange}
            onAddressIdChange={handleAddressIdChange}
            onUserAddressIdChange={handleUserAddressIdChange}
            onClientCreated={refreshClients}
            onUserAddressCreated={refreshUserAddresses}
            onClientAddressCreated={refreshClientAddresses}
          />
        )}

        {step === 1 && (
          <QuoteStepItems
            items={items}
            isReadonly={isCreate || isReadonly}
            totals={totals}
            availableTaxes={availableTaxes}
            taxById={taxById}
            isAdding={adding}
            onNameChange={handleNameChange}
            onQuantityChange={handleQuantityChange}
            onUnitPriceChange={handleUnitPriceChange}
            onTaxChange={handleTaxChange}
            onDescriptionChange={handleDescriptionChange}
            onOptionChange={handleOptionChange}
            onRemoveItem={handleRemoveItem}
            onAddItem={handleAddItem}
            onAddChildItem={handleAddChildItem}
            onSublineAdd={
              !isCreate && !isReadonly ? handleSublineAdd : undefined
            }
            onSublineChange={
              !isCreate && !isReadonly ? handleSublineChange : undefined
            }
            onSublineRemove={
              !isCreate && !isReadonly ? handleSublineRemove : undefined
            }
            onSaveLineAsTemplate={
              !isCreate && !isReadonly ? handleSaveLineAsTemplate : undefined
            }
            onAddItemFromTemplate={
              !isCreate && !isReadonly ? handleAddItemFromTemplate : undefined
            }
          />
        )}

        {step === 2 && <QuoteStepSummary />}
      </CardContent>

      <CardFooter className="justify-between border-t">
        <Button
          type="button"
          variant="outline"
          onClick={() => setStep((s) => Math.max(0, s - 1))}
          disabled={step === 0}
        >
          {t("prev")}
        </Button>

        <div className="flex gap-2">
          {step === 0 ? (
            <Button
              type="button"
              onClick={handleNextFromStep1}
              disabled={!canGoNextFromStep1}
            >
              {creating ? t("creating") : t("next")}
            </Button>
          ) : step < STEP_KEYS.length - 1 ? (
            <Button
              type="button"
              onClick={() =>
                setStep((s) => Math.min(STEP_KEYS.length - 1, s + 1))
              }
            >
              {t("next")}
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push("/quote")}
            >
              {t("finish")}
            </Button>
          )}
        </div>
      </CardFooter>

      <SaveTemplateDialog
        open={saveTemplateOpen}
        onOpenChange={setSaveTemplateOpen}
        defaultName={projectName}
        onSave={handleSaveQuoteAsTemplate}
      />

      <CreateScheduleDialog
        open={createScheduleOpen}
        onOpenChange={setCreateScheduleOpen}
        initialQuoteId={quoteId}
        lockQuote
      />
    </Card>
  );
}
