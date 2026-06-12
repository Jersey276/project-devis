"use client";

import { useEffect, useMemo, useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { createSchedule, listSchedules } from "@/lib/services/schedules";
import { listClients } from "@/lib/services/clients";
import { listQuotes } from "@/lib/services/quotes";
import type {
  BackendClient,
  BackendQuote,
  BackendScheduleSummary,
} from "@/types/backend";
import MonthPickerPopover from "@/components/schedule/month-picker-popover";

type QuoteOption = {
  quoteId: string;
  quoteName: string;
  label: string;
};

type CreateScheduleDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onCreated?: () => void;
  initialQuoteId?: string;
  lockQuote?: boolean;
};

function clientDisplayName(client: BackendClient | undefined): string {
  if (!client) return "Client inconnu";
  const fullName =
    `${client.first_name ?? ""} ${client.last_name ?? ""}`.trim();
  if (fullName) return fullName;
  if (client.company?.trim()) return client.company.trim();
  return client.client_id;
}

function buildQuoteOptions(
  quotes: BackendQuote[],
  clients: BackendClient[],
  blockedQuoteIds: Set<string>,
) {
  const clientsById = new Map(clients.map((c) => [c.client_id, c]));
  return quotes
    .filter((q) => !q.archived_at)
    .filter((q) => !blockedQuoteIds.has(q.quote_id))
    .sort((a, b) => a.name.localeCompare(b.name, "fr", { sensitivity: "base" }))
    .map((quote) => {
      const client = clientsById.get(quote.client_id);
      return {
        quoteId: quote.quote_id,
        quoteName: quote.name,
        label: `${quote.name} (${clientDisplayName(client)})`,
      };
    });
}

export default function CreateScheduleDialog({
  open,
  onOpenChange,
  onCreated,
  initialQuoteId,
  lockQuote = false,
}: CreateScheduleDialogProps) {
  const [quoteOptions, setQuoteOptions] = useState<QuoteOption[]>([]);
  const [creating, setCreating] = useState(false);
  const [errorMessage, setErrorMessage] = useState("");
  const [quoteId, setQuoteId] = useState("");
  const [name, setName] = useState("");
  const [startYear, setStartYear] = useState("");
  const [startMonthValue, setStartMonthValue] = useState("");
  const [startMonthPickerOpen, setStartMonthPickerOpen] = useState(false);
  const [durationMonths, setDurationMonths] = useState("");

  const effectiveQuoteId = lockQuote ? (initialQuoteId ?? "") : quoteId;

  const selectedQuote = useMemo(
    () => quoteOptions.find((q) => q.quoteId === effectiveQuoteId) ?? null,
    [effectiveQuoteId, quoteOptions],
  );

  const startMonth =
    startYear && startMonthValue ? `${startYear}-${startMonthValue}` : "";

  useEffect(() => {
    if (!open) return;
    let cancelled = false;

    Promise.all([listQuotes(), listClients(), listSchedules()]).then(
      ([quotesRes, clientsRes, schedulesRes]) => {
        if (cancelled) return;
        const quotes = Array.isArray(quotesRes.body.quotes)
          ? (quotesRes.body.quotes as BackendQuote[])
          : [];
        const clients = Array.isArray(clientsRes.body.clients)
          ? (clientsRes.body.clients as BackendClient[])
          : [];
        const schedules = Array.isArray(schedulesRes.body.schedules)
          ? (schedulesRes.body.schedules as BackendScheduleSummary[])
          : [];
        const blocked = new Set(
          schedules.filter((s) => s.status === "VALID").map((s) => s.quote_id),
        );

        setQuoteOptions(buildQuoteOptions(quotes, clients, blocked));

        if (lockQuote && initialQuoteId && blocked.has(initialQuoteId)) {
          setErrorMessage("Ce devis possède déjà un échéancier validé.");
        }
      },
    );

    return () => {
      cancelled = true;
    };
  }, [open, initialQuoteId, lockQuote]);

  async function onCreateSchedule() {
    setErrorMessage("");
    setCreating(true);
    const months = parseInt(durationMonths, 10);

    const { ok, body } = await createSchedule({
      quoteId: effectiveQuoteId,
      name: name.trim(),
      startMonth: startMonth.trim(),
      durationMonths: Number.isInteger(months) && months > 0 ? months : 0,
    });

    if (!ok || !body.success) {
      setErrorMessage((body.message as string) ?? "Données invalides.");
      setCreating(false);
      return;
    }

    setName("");
    setStartYear("");
    setStartMonthValue("");
    setDurationMonths("");
    setCreating(false);
    onOpenChange(false);
    onCreated?.();
  }

  const selectedLabel = selectedQuote?.label ?? effectiveQuoteId;

  function handleOpenChange(nextOpen: boolean) {
    if (!nextOpen) {
      setErrorMessage("");
      setName("");
      setStartYear("");
      setStartMonthValue("");
      setDurationMonths("");
      setStartMonthPickerOpen(false);
      if (!lockQuote) setQuoteId("");
    }
    onOpenChange(nextOpen);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Nouvel échéancier</DialogTitle>
        </DialogHeader>

        <div className="space-y-3">
          <Field>
            <FieldLabel htmlFor="schedule-quote-id">ID devis</FieldLabel>
            {lockQuote ? (
              <Input
                id="schedule-quote-id"
                name="quote_id"
                value={selectedLabel}
                readOnly
              />
            ) : (
              <Combobox
                items={quoteOptions}
                value={selectedQuote}
                onValueChange={(item: QuoteOption | null) =>
                  setQuoteId(item?.quoteId ?? "")
                }
                itemToStringLabel={(item: QuoteOption) => item.label}
              >
                <ComboboxInput
                  id="schedule-quote-id"
                  name="quote_id"
                  placeholder="Sélectionner un devis"
                />
                <ComboboxContent>
                  <ComboboxEmpty>Aucun devis disponible.</ComboboxEmpty>
                  <ComboboxList>
                    {(item: QuoteOption) => (
                      <ComboboxItem key={item.quoteId} value={item}>
                        {item.label}
                      </ComboboxItem>
                    )}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
            )}
          </Field>

          <Field>
            <FieldLabel htmlFor="schedule-name">Nom</FieldLabel>
            <Input
              id="schedule-name"
              name="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="schedule-start-month">
              Mois de début
            </FieldLabel>
            <MonthPickerPopover
              id="schedule-start-month"
              name="start_month"
              value={startMonth}
              startYear={startYear}
              startMonthValue={startMonthValue}
              open={startMonthPickerOpen}
              onOpenChange={setStartMonthPickerOpen}
              onStartYearChange={setStartYear}
              onStartMonthChange={setStartMonthValue}
            />
          </Field>

          <Field>
            <FieldLabel htmlFor="schedule-duration">Durée (mois)</FieldLabel>
            <Input
              id="schedule-duration"
              name="duration_months"
              inputMode="numeric"
              value={durationMonths}
              onChange={(e) => setDurationMonths(e.target.value)}
            />
          </Field>

          <FieldError>{errorMessage}</FieldError>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            Annuler
          </Button>
          <Button
            type="button"
            onClick={onCreateSchedule}
            disabled={creating || !effectiveQuoteId || !startMonth}
          >
            {creating ? "Création…" : "Créer"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
