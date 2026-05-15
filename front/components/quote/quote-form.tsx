"use client";

import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { useRouter, useSearchParams } from "next/navigation";
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
import { DownloadIcon, Loader2Icon } from "lucide-react";
import QuoteStepBasicInfo from "@/components/quote/steps/quote-step-basic-info";
import QuoteStepItems, {
  type QuoteItemRow as RenderedRow,
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
import { listClients } from "@/lib/services/clients";
import { listAddresses } from "@/lib/services/addresses";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { useMode } from "@/lib/mode-context";
import {
  QUOTE_STATE_LABEL,
  type BackendAddress,
  type BackendClient,
  type BackendQuote,
  type BackendQuoteLine,
  type BackendQuoteState,
  type BackendTax,
} from "@/types/backend";

type FormItem = RenderedRow & { position: number };

type QuoteFormProps = {
  quoteId?: string;
};

const STEP_LABELS = [
  "Informations de base",
  "Éléments du devis",
  "Résumé",
] as const;

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
    name: line.name,
    quantity: Number(line.quantity),
    unitPriceEuros: line.unit_price / 100,
    position: line.position,
    taxId: line.tax_id ?? null,
    saveStatus: "idle",
  };
}

function lineDraftFromRow(row: FormItem): LineDraft {
  return {
    type: "simple",
    name: row.name,
    quantity: row.quantity,
    unitPriceEuros: row.unitPriceEuros,
    position: row.position,
    taxId: row.taxId,
  };
}

export default function QuoteForm({ quoteId }: QuoteFormProps) {
  const router = useRouter();
  const searchParams = useSearchParams();
  const { isCustomer } = useMode();
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
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [addresses, setAddresses] = useState<BackendAddress[]>([]);
  const [items, setItems] = useState<FormItem[]>([]);
  const [availableTaxes, setAvailableTaxes] = useState<BackendTax[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [creating, setCreating] = useState(false);
  const [adding, setAdding] = useState(false);
  const [quoteState, setQuoteState] =
    useState<BackendQuoteState>("draft");
  const [transitioning, setTransitioning] = useState(false);
  const [isExporting, setIsExporting] = useState(false);

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
        toast.error((body.message as string) ?? "Devis introuvable.");
        return;
      }
      const fetchedQuote = body.quote as BackendQuote;
      const fetchedLines = (body.lines ?? []) as BackendQuoteLine[];
      setProjectName(fetchedQuote.name);
      setQuoteState(fetchedQuote.state ?? "draft");
      setClientId(fetchedQuote.client_id ?? "");
      setAddressId(fetchedQuote.address_id ?? null);
      const sorted = [...fetchedLines].sort(
        (a, b) => a.position - b.position,
      );
      setItems(sorted.map(rowFromBackendLine));
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [quoteId]);

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

  // Load taxes available for the current user (resolved from their first
  // address). Empty when the user has no address or no taxes apply.
  useEffect(() => {
    let cancelled = false;
    listAvailableTaxesForUser().then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.taxes)) {
        setAvailableTaxes(body.taxes as BackendTax[]);
      } else {
        setAvailableTaxes([]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

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
      for (const t of lineTimers.values()) clearTimeout(t);
      lineTimers.clear();
      for (const t of savedTimers.values()) clearTimeout(t);
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

  const totals = useMemo(() => {
    const ht = items.reduce(
      (acc, item) => acc + item.quantity * item.unitPriceEuros,
      0,
    );
    const breakdown = new Map<number, { tax: BackendTax; amount: number }>();
    for (const item of items) {
      if (item.taxId == null) continue;
      const tax = taxById.get(item.taxId);
      if (!tax) continue;
      const lineHT = item.quantity * item.unitPriceEuros;
      const taxAmount = (lineHT * Number(tax.rate)) / 100;
      const cur = breakdown.get(tax.id);
      breakdown.set(tax.id, {
        tax,
        amount: (cur?.amount ?? 0) + taxAmount,
      });
    }
    const sortedBreakdown = Array.from(breakdown.values()).sort(
      (a, b) => Number(a.tax.rate) - Number(b.tax.rate),
    );
    const ttc = ht + sortedBreakdown.reduce((acc, b) => acc + b.amount, 0);
    return { ht, breakdown: sortedBreakdown, ttc };
  }, [items, taxById]);

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
            (body.message as string) ??
              "Échec de l'enregistrement du nom.",
          );
        }
      }, SAVE_DEBOUNCE_MS);
    },
    [isReadonly, quoteId],
  );

  const handleNextFromStep1 = useCallback(async () => {
    const trimmed = projectName.trim();
    const localErrors: FieldErrors = {};
    if (trimmed.length === 0) {
      localErrors.name = ["Ce champ est requis."];
    }
    if (!clientId) {
      localErrors.client_id = ["Veuillez sélectionner un client."];
    }
    if (!addressId) {
      localErrors.address_id = ["Veuillez sélectionner une adresse."];
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
        });
        if (ok && body.success) {
          const newId = body.quote_id as string;
          router.replace(`/quote/${newId}?step=2`);
          return;
        }
        if (status === 422 && Array.isArray(body.field_errors)) {
          setErrors(fieldErrorsFromBody(body));
        } else {
          toast.error(
            (body.message as string) ?? "Une erreur est survenue.",
          );
        }
      } catch {
        toast.error("Une erreur est survenue.");
      } finally {
        setCreating(false);
      }
      return;
    }
    setStep(1);
  }, [addressId, clientId, isCreate, projectName, router]);

  // ────────────────────────────────────────────────────────────
  // Step 2 handlers

  const setRow = useCallback(
    (lineId: string, patch: Partial<FormItem>) => {
      setItems((prev) =>
        prev.map((row) =>
          row.lineId === lineId ? { ...row, ...patch } : row,
        ),
      );
    },
    [],
  );

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
          const t = setTimeout(() => {
            savedIndicatorTimersRef.current.delete(lineId);
            setRow(lineId, { saveStatus: "idle" });
          }, SAVED_INDICATOR_MS);
          savedIndicatorTimersRef.current.set(lineId, t);
        } else {
          setRow(lineId, { saveStatus: "error" });
          toast.error(
            (body.message as string) ??
              "Échec d'enregistrement de la ligne.",
          );
        }
      }, SAVE_DEBOUNCE_MS);

      lineTimersRef.current.set(lineId, timer);
    },
    [isReadonly, quoteId, setRow],
  );

  const handleDrop = useCallback(async () => {
    if (!quoteId || transitioning) return;
    setTransitioning(true);
    try {
      const { ok, body } = await dropQuote(quoteId);
      if (ok && body.success) {
        setQuoteState("drop");
      } else {
        toast.error(
          (body.message as string) ?? "Impossible d'abandonner le devis.",
        );
      }
    } finally {
      setTransitioning(false);
    }
  }, [quoteId, transitioning]);

  const handleContinue = useCallback(async () => {
    if (!quoteId || transitioning) return;
    setTransitioning(true);
    try {
      const { ok, body } = await continueQuote(quoteId);
      if (ok && body.success) {
        setQuoteState("draft");
      } else {
        toast.error(
          (body.message as string) ?? "Impossible de réactiver le devis.",
        );
      }
    } finally {
      setTransitioning(false);
    }
  }, [quoteId, transitioning]);

  const handleExport = useCallback(async () => {
    if (!quoteId || isExporting) return;
    setIsExporting(true);
    try {
      await exportQuotePdf(quoteId);
    } catch (err) {
      console.error("export quote pdf failed", err);
      toast.error("Échec de l'export.");
    } finally {
      setIsExporting(false);
    }
  }, [quoteId, isExporting]);

  const handleAddItem = useCallback(async () => {
    if (!quoteId || adding) return;
    setAdding(true);
    try {
      const draft: LineDraft = {
        type: "simple",
        name: "",
        quantity: 1,
        unitPriceEuros: 0,
        position: itemsRef.current.length,
        taxId: defaultTaxId,
      };
      const { ok, body } = await createLine(quoteId, draft);
      if (ok && body.success) {
        const newLineId = body.line_id as string;
        setItems((prev) => [
          ...prev,
          {
            lineId: newLineId,
            name: "",
            quantity: 1,
            unitPriceEuros: 0,
            position: prev.length,
            taxId: defaultTaxId,
            saveStatus: "idle",
          },
        ]);
      } else {
        toast.error(
          (body.message as string) ?? "Impossible d'ajouter la ligne.",
        );
      }
    } finally {
      setAdding(false);
    }
  }, [adding, quoteId, defaultTaxId]);

  const handleRemoveItem = useCallback(
    async (lineId: string) => {
      if (!quoteId) return;
      if (itemsRef.current.length <= 1) return;
      const t = lineTimersRef.current.get(lineId);
      if (t) {
        clearTimeout(t);
        lineTimersRef.current.delete(lineId);
      }
      const previous = itemsRef.current;
      setItems((prev) => prev.filter((row) => row.lineId !== lineId));
      const { ok, body } = await deleteLine(quoteId, lineId);
      if (!ok || !body.success) {
        toast.error(
          (body.message as string) ?? "Impossible de supprimer la ligne.",
        );
        setItems(previous);
      }
    },
    [quoteId],
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
      void updateQuote(quoteId, { name, clientId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error(
            (body.message as string) ?? "Échec de l'enregistrement du client.",
          );
        }
      });
    },
    [isReadonly, quoteId],
  );

  const handleAddressIdChange = useCallback(
    (value: number | null) => {
      setAddressId(value);
      setErrors((prev) => ({ ...prev, address_id: [] }));
      if (!quoteId || isReadonly || value == null) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, addressId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error(
            (body.message as string) ?? "Échec de l'enregistrement de l'adresse.",
          );
        }
      });
    },
    [isReadonly, quoteId],
  );

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
    projectName.trim().length > 0 && !!clientId && !!addressId && !creating;

  const showDropButton =
    !isCustomer && !isCreate && (quoteState === "draft" || quoteState === "sent");
  const showContinueButton = !isCustomer && !isCreate && quoteState === "drop";

  return (
    <Card data-quote-state={quoteState}>
      <CardHeader className="flex flex-row items-start justify-between gap-4">
        <div className="space-y-1.5">
          <CardTitle className="flex items-center gap-2">
            {isCreate ? "Création d'un devis" : `Devis ${projectName || "…"}`}
            {!isCreate && (
              <Badge
                data-slot="quote-state-badge"
                variant={STATE_BADGE_VARIANT[quoteState]}
              >
                {QUOTE_STATE_LABEL[quoteState]}
              </Badge>
            )}
          </CardTitle>
          <CardDescription>
            {isCreate
              ? "Complète les étapes pour générer un nouveau devis."
              : "Modification du devis."}
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
            Exporter
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
                Abandonner
              </Button>
            </AlertDialogTrigger>
            <AlertDialogContent>
              <AlertDialogHeader>
                <AlertDialogTitle>Abandonner ce devis ?</AlertDialogTitle>
                <AlertDialogDescription>
                  Le devis ne pourra plus être modifié. Vous pourrez le
                  réactiver via le bouton Continuer.
                </AlertDialogDescription>
              </AlertDialogHeader>
              <AlertDialogFooter>
                <AlertDialogCancel>Annuler</AlertDialogCancel>
                <AlertDialogAction
                  variant="destructive"
                  onClick={handleDrop}
                >
                  Confirmer
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
            Continuer
          </Button>
        )}
      </CardHeader>

      <CardContent className="space-y-6">
        <div className="grid gap-2 sm:grid-cols-3">
          {STEP_LABELS.map((label, index) => (
            <Button
              key={label}
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
              Étape {index + 1} · {label}
            </Button>
          ))}
        </div>

        {step === 0 && (
          <QuoteStepBasicInfo
            projectName={projectName}
            clientId={clientId}
            addressId={addressId}
            isReadonly={isReadonly}
            clients={clients}
            addresses={addresses}
            nameErrors={errors.name}
            clientErrors={errors.client_id}
            addressErrors={errors.address_id}
            onProjectNameChange={handleProjectNameChange}
            onClientIdChange={handleClientIdChange}
            onAddressIdChange={handleAddressIdChange}
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
            onRemoveItem={handleRemoveItem}
            onAddItem={handleAddItem}
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
          Précédent
        </Button>

        <div className="flex gap-2">
          {step === 0 ? (
            <Button
              type="button"
              onClick={handleNextFromStep1}
              disabled={!canGoNextFromStep1}
            >
              {creating ? "Création…" : "Suivant"}
            </Button>
          ) : step < STEP_LABELS.length - 1 ? (
            <Button
              type="button"
              onClick={() =>
                setStep((s) => Math.min(STEP_LABELS.length - 1, s + 1))
              }
            >
              Suivant
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push("/quote")}
            >
              Terminer
            </Button>
          )}
        </div>
      </CardFooter>
    </Card>
  );
}
