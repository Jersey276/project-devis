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
  DownloadIcon,
  BookmarkIcon,
  Loader2Icon,
  CalendarIcon,
} from "lucide-react";
import QuoteStepBasicInfo from "@/components/quote/steps/quote-step-basic-info";
import QuoteStepItems from "@/components/quote/steps/quote-step-items";
import QuoteStepSummary from "@/components/quote/steps/quote-step-summary";
import {
  createLine,
  createQuote,
  getQuote,
  updateQuote,
} from "@/lib/services/quotes";
import { exportQuotePdf } from "@/lib/services/export";
import GenerateInvoiceFromQuoteButton from "@/components/invoice/generate-invoice-from-quote-button";
import QuoteStateDropdown from "@/components/quote/quote-state-dropdown";
import { listTemplateLines } from "@/lib/services/templates";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { useMode } from "@/lib/mode-context";
import {
  type BackendQuote,
  type BackendQuoteLine,
  type BackendQuoteState,
  type BackendTemplateLine,
} from "@/types/backend";
import {
  computeTotals,
  rowFromBackendLine,
  type FormItem,
  useQuoteLines,
} from "@/hooks/use-quote-lines";
import { useQuoteReferenceData } from "@/hooks/use-quote-reference-data";
import SaveTemplateDialog from "@/components/template/save-template-dialog";
import CreateScheduleDialog from "@/components/schedule/create-schedule-dialog";

type QuoteFormProps = {
  quoteId?: string;
};

const STEP_KEYS = ["basicInfo", "items"] as const;
const SAVE_DEBOUNCE_MS = 600;

const STATE_BADGE_VARIANT: Record<
  BackendQuoteState,
  "default" | "secondary" | "destructive"
> = {
  draft: "secondary",
  negociation: "default",
  validated: "default",
  drop: "destructive",
};

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

  const templateIdFromQuery = useMemo(
    () => searchParams.get("template") ?? null,
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [],
  );

  const [step, setStep] = useState(initialStep);
  const [loading, setLoading] = useState(!isCreate);
  const [notFound, setNotFound] = useState(false);
  const [projectName, setProjectName] = useState("");
  const [clientId, setClientId] = useState("");
  const [addressId, setAddressId] = useState<number | null>(null);
  const [userAddressId, setUserAddressId] = useState<number | null>(null);
  const [items, setItems] = useState<FormItem[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [creating, setCreating] = useState(false);
  const [quoteState, setQuoteState] = useState<BackendQuoteState>("draft");
  const [isExporting, setIsExporting] = useState(false);
  const [saveTemplateOpen, setSaveTemplateOpen] = useState(false);
  const [createScheduleOpen, setCreateScheduleOpen] = useState(false);

  const isReadonly =
    isCustomer || quoteState === "validated" || quoteState === "drop";

  const projectNameRef = useRef(projectName);
  useEffect(() => { projectNameRef.current = projectName; }, [projectName]);
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
    return () => { cancelled = true; };
  }, [quoteId, t]);

  useEffect(() => {
    if (notFound) router.replace("/quote");
  }, [notFound, router]);

  useEffect(() => {
    return () => {
      if (nameTimerRef.current) clearTimeout(nameTimerRef.current);
    };
  }, []);

  const {
    clients,
    userId,
    userAddresses,
    addresses,
    availableTaxes,
    taxById,
    defaultTaxId,
    refreshClients,
    refreshUserAddresses,
    refreshClientAddresses,
  } = useQuoteReferenceData({ clientId, userAddressId, isCustomer, loading, items });

  const {
    adding,
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
  } = useQuoteLines({ quoteId, isReadonly, defaultTaxId, items, setItems });

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
          toast.error((body.message as string) ?? t("errors.nameSaveFailedToast"));
        }
      }, SAVE_DEBOUNCE_MS);
    },
    [isReadonly, quoteId, t],
  );

  const handleNextFromStep1 = useCallback(async () => {
    const trimmed = projectName.trim();
    const localErrors: FieldErrors = {};
    if (trimmed.length === 0) localErrors.name = [t("errors.requiredField")];
    if (!clientId) localErrors.client_id = [t("errors.selectClient")];
    if (!addressId) localErrors.address_id = [t("errors.selectAddress")];
    if (!userAddressId) localErrors.user_address_id = [t("errors.selectUserAddress")];
    if (Object.keys(localErrors).length > 0) { setErrors(localErrors); return; }

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
              const sorted = (linesRes.body.lines as BackendTemplateLine[]).sort(
                (a, b) => a.position - b.position,
              );
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
  }, [addressId, clientId, isCreate, projectName, router, t, tCommon, templateIdFromQuery, userAddressId]);

  const handleClientIdChange = useCallback(
    (value: string) => {
      setClientId(value);
      setAddressId(null);
      setErrors((prev) => ({ ...prev, client_id: [], address_id: [] }));
      if (!quoteId || isReadonly || !value) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, clientId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.clientSaveFailedToast"));
        }
      });
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
      void updateQuote(quoteId, { name, addressId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.addressSaveFailedToast"));
        }
      });
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
      void updateQuote(quoteId, { name, userAddressId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.userAddressSaveFailedToast"));
        }
      });
    },
    [isReadonly, quoteId, t],
  );

  const handleExport = useCallback(async () => {
    if (!quoteId || isExporting) return;
    setIsExporting(true);
    try {
      await exportQuotePdf(quoteId);
    } catch (err) {
      console.error("export quote pdf failed", err);
      toast.error(t("exportFailedToast"));
    } finally { setIsExporting(false); }
  }, [quoteId, isExporting, t]);

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
        {!isCreate && quoteId ? (
          <GenerateInvoiceFromQuoteButton
            quoteId={quoteId}
            validated={quoteState === "validated"}
            onError={(message) => toast.error(message)}
          />
        ) : null}
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
        {!isCustomer && !isCreate && quoteId && (
          <QuoteStateDropdown
            quoteId={quoteId}
            state={quoteState}
            onChanged={(next) => setQuoteState(next)}
            onError={(message) => toast.error(message)}
          />
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
              onClick={() => { if (!isCreate) setStep(index); }}
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
            onAddFeeItem={!isCreate && !isReadonly ? handleAddFeeItem : undefined}
            onAddFeeSubline={!isCreate && !isReadonly ? handleAddFeeSubline : undefined}
            onSublineAdd={!isCreate && !isReadonly ? handleSublineAdd : undefined}
            onSublineChange={!isCreate && !isReadonly ? handleSublineChange : undefined}
            onSublineRemove={!isCreate && !isReadonly ? handleSublineRemove : undefined}
            onSaveLineAsTemplate={!isCreate && !isReadonly ? handleSaveLineAsTemplate : undefined}
            onAddItemFromTemplate={!isCreate && !isReadonly ? handleAddItemFromTemplate : undefined}
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
              onClick={() => setStep((s) => Math.min(STEP_KEYS.length - 1, s + 1))}
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
