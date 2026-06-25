"use client";

import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import {
  Card,
  CardContent,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Loader2Icon } from "lucide-react";
import QuoteStepBasicInfo from "@/components/quote/steps/quote-step-basic-info";
import QuoteStepItems from "@/components/quote/steps/quote-step-items";
import QuoteStepSummary from "@/components/quote/steps/quote-step-summary";
import {
  createLine,
  createQuote,
  getMyQuote,
  getQuote,
  negociateQuote,
  updateMyQuoteAddress,
  updateQuote,
} from "@/lib/services/quotes";
import { exportQuotePdf } from "@/lib/services/export";
import { listTemplateLines } from "@/lib/services/templates";
import { apiFetch, fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
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
import QuoteFormDialogs from "@/components/quote/quote-form-dialogs";
import QuoteFormFooter from "@/components/quote/quote-form-footer";
import QuoteFormHeader from "@/components/quote/quote-form-header";

type QuoteFormProps = {
  quoteId?: string;
};

const STEP_KEYS = ["basicInfo", "items"] as const;
const SAVE_DEBOUNCE_MS = 600;

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
  const [validUntil, setValidUntil] = useState("");
  const [paymentTerms, setPaymentTerms] = useState("");
  const [items, setItems] = useState<FormItem[]>([]);
  const [errors, setErrors] = useState<FieldErrors>({});
  const [creating, setCreating] = useState(false);
  const [quoteState, setQuoteState] = useState<BackendQuoteState>("draft");
  const [isExporting, setIsExporting] = useState(false);
  const [saveTemplateOpen, setSaveTemplateOpen] = useState(false);
  const [createScheduleOpen, setCreateScheduleOpen] = useState(false);
  const [commentSidebarOpen, setCommentSidebarOpen] = useState(false);
  const [commentLineId, setCommentLineId] = useState("");
  const [commentLineName, setCommentLineName] = useState("");
  const [currentUserName, setCurrentUserName] = useState("");
  const [customerUserId, setCustomerUserId] = useState("");

  const isReadonly =
    isCustomer || quoteState === "validated" || quoteState === "drop";

  const projectNameRef = useRef(projectName);
  useEffect(() => { projectNameRef.current = projectName; }, [projectName]);
  const nameTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Initial fetch (edit mode only)
  useEffect(() => {
    if (!quoteId) return;
    let cancelled = false;
    const fetch = isCustomer ? getMyQuote(quoteId) : getQuote(quoteId);
    fetch.then(({ ok, body }) => {
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
      setValidUntil(fetchedQuote.valid_until ?? "");
      setPaymentTerms(fetchedQuote.payment_terms ?? "");
      const sorted = [...fetchedLines].sort((a, b) => a.position - b.position);
      setItems(sorted.map(rowFromBackendLine));
      setLoading(false);
    });
    return () => { cancelled = true; };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [quoteId, t]);

  useEffect(() => {
    if (notFound) router.replace("/quote");
  }, [notFound, router]);

  // Fetch current user display name for comment authorship.
  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (isCustomer) {
        const [clientsRes, meRes] = await Promise.all([
          apiFetch("/api/users/clients/me"),
          apiFetch("/api/auth/me"),
        ]);
        if (cancelled) return;
        if (clientsRes.ok && Array.isArray(clientsRes.body.clients) && clientsRes.body.clients.length > 0) {
          const cl = clientsRes.body.clients[0] as { first_name: string; last_name: string };
          setCurrentUserName(`${cl.first_name} ${cl.last_name}`.trim());
        }
        if (meRes.ok && meRes.body.auth) {
          setCustomerUserId((meRes.body.auth as { user_id: string }).user_id);
        }
      } else {
        const { ok, body } = await apiFetch("/api/users/me");
        if (cancelled) return;
        if (ok && body.success && body.user) {
          const u = body.user as { company?: string; email?: string };
          setCurrentUserName(u.company || u.email || "");
        }
      }
    })();
    return () => { cancelled = true; };
  }, [isCustomer]);

  useEffect(() => {
    return () => {
      if (nameTimerRef.current) clearTimeout(nameTimerRef.current);
    };
  }, []);

  const {
    clients,
    userId,
    myClientId,
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
  // Helpers

  // Called after any updateQuote that implicitly reverted a negociation quote
  // back to draft. Syncs local state and offers to resend to the client.
  const handleRevertedToDraft = useCallback(() => {
    setQuoteState("draft");
    if (!quoteId) return;
    toast(t("errors.revertedToDraftToast"), {
      action: {
        label: t("errors.resendToClient"),
        onClick: () => {
          void negociateQuote(quoteId).then(({ ok, body }) => {
            if (ok && body.success) {
              setQuoteState("negociation");
            } else {
              toast.error((body.message as string) ?? t("errors.nameSaveFailedToast"));
            }
          });
        },
      },
    });
  }, [quoteId, t]);

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
        const wasNegociation = quoteState === "negociation";
        const { ok, body } = await updateQuote(quoteId, { name: current });
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.nameSaveFailedToast"));
        } else if (wasNegociation) {
          handleRevertedToDraft();
        }
      }, SAVE_DEBOUNCE_MS);
    },
    [handleRevertedToDraft, isReadonly, quoteId, quoteState, t],
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
        if (status === 422) {
          const parsed = fieldErrorsFromBody(body);
          if (Object.keys(parsed).length > 0) {
            setErrors(parsed);
          } else {
            toast.error((body.message as string) ?? tCommon("errors.generic"));
          }
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
      const wasNegociation = quoteState === "negociation";
      void updateQuote(quoteId, { name, clientId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.clientSaveFailedToast"));
        } else if (wasNegociation) {
          handleRevertedToDraft();
        }
      });
    },
    [handleRevertedToDraft, isReadonly, quoteId, quoteState, t],
  );

  const handleAddressIdChange = useCallback(
    (value: number | null) => {
      setAddressId(value);
      setErrors((prev) => ({ ...prev, address_id: [] }));
      if (!quoteId || value == null) return;
      if (isCustomer) {
        void updateMyQuoteAddress(quoteId, value, myClientId || undefined).then(({ ok, body }) => {
          if (!ok || !body.success) {
            toast.error((body.message as string) ?? t("errors.addressSaveFailedToast"));
          }
        });
        return;
      }
      if (isReadonly) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      const wasNegociation = quoteState === "negociation";
      void updateQuote(quoteId, { name, addressId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.addressSaveFailedToast"));
        } else if (wasNegociation) {
          handleRevertedToDraft();
        }
      });
    },
    [handleRevertedToDraft, isCustomer, isReadonly, myClientId, quoteId, quoteState, t],
  );

  const handleUserAddressIdChange = useCallback(
    (value: number | null) => {
      setUserAddressId(value);
      setErrors((prev) => ({ ...prev, user_address_id: [] }));
      if (!quoteId || isReadonly || value == null) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      const wasNegociation = quoteState === "negociation";
      void updateQuote(quoteId, { name, userAddressId: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.userAddressSaveFailedToast"));
        } else if (wasNegociation) {
          handleRevertedToDraft();
        }
      });
    },
    [handleRevertedToDraft, isReadonly, quoteId, quoteState, t],
  );

  const handleValidUntilChange = useCallback(
    (value: string) => {
      setValidUntil(value);
      if (!quoteId || isReadonly) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, validUntil: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.nameSaveFailedToast"));
        }
      });
    },
    [isReadonly, quoteId, t],
  );

  const handlePaymentTermsChange = useCallback(
    (value: string) => {
      setPaymentTerms(value);
      if (!quoteId || isReadonly) return;
      const name = projectNameRef.current.trim();
      if (name.length === 0) return;
      void updateQuote(quoteId, { name, paymentTerms: value }).then(({ ok, body }) => {
        if (!ok || !body.success) {
          toast.error((body.message as string) ?? t("errors.nameSaveFailedToast"));
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
      <QuoteFormHeader
        quoteId={quoteId}
        projectName={projectName}
        quoteState={quoteState}
        isCreate={isCreate}
        isReadonly={isReadonly}
        isCustomer={isCustomer}
        isExporting={isExporting}
        onExport={handleExport}
        onSaveTemplate={() => setSaveTemplateOpen(true)}
        onCreateSchedule={() => setCreateScheduleOpen(true)}
        onStateChanged={(next) => setQuoteState(next)}
      />

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
            validUntil={validUntil}
            paymentTerms={paymentTerms}
            isReadonly={isReadonly}
            customerAddressEditable={isCustomer}
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
            onValidUntilChange={handleValidUntilChange}
            onPaymentTermsChange={handlePaymentTermsChange}
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
            onOpenComments={!isCreate && quoteId ? (lineId, lineName) => {
              setCommentLineId(lineId);
              setCommentLineName(lineName);
              setCommentSidebarOpen(true);
            } : undefined}
          />
        )}

        {step === 2 && <QuoteStepSummary />}
      </CardContent>

      <QuoteFormFooter
        step={step}
        stepCount={STEP_KEYS.length}
        creating={creating}
        canGoNextFromStep1={canGoNextFromStep1}
        onPrev={() => setStep((s) => Math.max(0, s - 1))}
        onNextFromStep1={handleNextFromStep1}
        onNextStep={() => setStep((s) => Math.min(STEP_KEYS.length - 1, s + 1))}
        onFinish={() => router.push("/quote")}
      />

      <QuoteFormDialogs
        quoteId={quoteId}
        projectName={projectName}
        saveTemplateOpen={saveTemplateOpen}
        onSaveTemplateOpenChange={setSaveTemplateOpen}
        onSaveQuoteAsTemplate={handleSaveQuoteAsTemplate}
        createScheduleOpen={createScheduleOpen}
        onCreateScheduleOpenChange={setCreateScheduleOpen}
        commentSidebarOpen={commentSidebarOpen}
        onCommentSidebarOpenChange={setCommentSidebarOpen}
        commentLineId={commentLineId}
        commentLineName={commentLineName}
        currentUserId={isCustomer ? customerUserId : userId}
        currentUserName={currentUserName}
      />
    </Card>
  );
}
