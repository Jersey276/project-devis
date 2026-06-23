"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ChevronDownIcon, ChevronRightIcon, UnlinkIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { toast } from "sonner";
import { listQuotes } from "@/lib/services/quotes";
import {
  addQuoteToProject,
  removeQuoteFromProject,
} from "@/lib/services/projects";
import { formatEurosFromCents } from "@/lib/utils";
import type {
  BackendInvoiceSummary,
  BackendProjectQuoteRow,
  BackendQuote,
  BackendScheduleSummary,
} from "@/types/backend";

const QUOTE_STATE_LABELS: Record<string, string> = {
  draft: "Brouillon",
  negociation: "Négociation",
  validated: "Validé",
  drop: "Abandonné",
  sent: "Envoyé",
};

const QUOTE_STATE_VARIANTS: Record<
  string,
  "default" | "secondary" | "outline" | "destructive"
> = {
  draft: "outline",
  negociation: "default",
  validated: "secondary",
  drop: "destructive",
  sent: "default",
};

const SCHEDULE_STATUS_LABELS: Record<string, string> = {
  DRAFT: "Brouillon",
  NEGOCIATE: "Négociation",
  VALID: "Validé",
  DENIED: "Refusé",
};

const INVOICE_STATUS_LABELS: Record<string, string> = {
  DRAFT: "Brouillon",
  ISSUED: "Émise",
  PAID: "Payée",
  CANCELLED: "Annulée",
};

const INVOICE_STATUS_VARIANTS: Record<
  string,
  "default" | "secondary" | "outline" | "destructive"
> = {
  DRAFT: "outline",
  ISSUED: "default",
  PAID: "secondary",
  CANCELLED: "destructive",
};

type Props = {
  projectId: string;
  quotes: BackendProjectQuoteRow[];
  onChanged: () => void;
};

function ScheduleSubTable({
  schedules,
}: {
  schedules: BackendScheduleSummary[];
}) {
  if (!schedules?.length)
    return <p className="text-xs text-muted-foreground">Aucun échéancier.</p>;
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="text-xs">Nom</TableHead>
          <TableHead className="text-xs">Statut</TableHead>
          <TableHead className="text-xs">Début</TableHead>
          <TableHead className="text-xs">Durée</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {schedules.map((s) => (
          <TableRow key={s.schedule_id}>
            <TableCell className="text-xs">
              <Link
                href={`/schedule/${s.schedule_id}`}
                className="hover:underline"
              >
                {s.name}
              </Link>
            </TableCell>
            <TableCell className="text-xs">
              {SCHEDULE_STATUS_LABELS[s.status] ?? s.status}
            </TableCell>
            <TableCell className="text-xs">{s.start_month}</TableCell>
            <TableCell className="text-xs">{s.duration_months} mois</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

function InvoiceSubTable({ invoices }: { invoices: BackendInvoiceSummary[] }) {
  if (!invoices?.length)
    return <p className="text-xs text-muted-foreground">Aucune facture.</p>;
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="text-xs">N°</TableHead>
          <TableHead className="text-xs">Statut</TableHead>
          <TableHead className="text-xs">Émission</TableHead>
          <TableHead className="text-xs">Total TTC</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {invoices.map((inv) => (
          <TableRow key={inv.invoice_id}>
            <TableCell className="text-xs">
              <Link
                href={`/invoice/${inv.invoice_id}`}
                className="hover:underline"
              >
                {inv.invoice_number || "Brouillon"}
              </Link>
            </TableCell>
            <TableCell className="text-xs">
              <Badge
                variant={INVOICE_STATUS_VARIANTS[inv.status] ?? "outline"}
                className="text-xs"
              >
                {INVOICE_STATUS_LABELS[inv.status] ?? inv.status}
              </Badge>
            </TableCell>
            <TableCell className="text-xs">
              {inv.issued_at
                ? new Date(inv.issued_at).toLocaleDateString("fr-FR")
                : "—"}
            </TableCell>
            <TableCell className="text-xs">
              {formatEurosFromCents(inv.total_ttc_cents)}
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

export default function ProjectQuotesTable({
  projectId,
  quotes,
  onChanged,
}: Props) {
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [availableQuotes, setAvailableQuotes] = useState<BackendQuote[]>([]);
  const [addQuoteId, setAddQuoteId] = useState("");
  const [busy, setBusy] = useState(false);

  const linkedIds = new Set(quotes.map((q) => q.quote_id));

  useEffect(() => {
    listQuotes("page_size=200").then(({ ok, body }) => {
      if (ok && Array.isArray(body.quotes)) {
        setAvailableQuotes(
          (body.quotes as BackendQuote[]).filter(
            (q) => !linkedIds.has(q.quote_id),
          ),
        );
      }
    });
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [quotes]);

  function toggle(id: string) {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  async function handleAdd() {
    if (!addQuoteId) return;
    setBusy(true);
    const { ok, body } = await addQuoteToProject(projectId, addQuoteId);
    setBusy(false);
    if (!ok) {
      toast.error(body?.message ?? "Impossible d'ajouter le devis.");
      return;
    }
    setAddQuoteId("");
    onChanged();
  }

  async function handleRemove(quoteId: string) {
    setBusy(true);
    const { ok, body } = await removeQuoteFromProject(projectId, quoteId);
    setBusy(false);
    if (!ok) {
      toast.error(body?.message ?? "Impossible de retirer le devis.");
      return;
    }
    onChanged();
  }

  return (
    <div className="flex flex-col gap-4">
      {/* Add quote row */}
      <div className="flex items-center gap-2">
        <div className="w-64">
          <Combobox value={addQuoteId} onValueChange={(v) => setAddQuoteId(v ?? "")}>
            <ComboboxInput placeholder="Ajouter un devis…" />
            <ComboboxContent>
              <ComboboxList>
                <ComboboxEmpty>Aucun devis disponible.</ComboboxEmpty>
                {availableQuotes.map((q) => (
                  <ComboboxItem key={q.quote_id} value={q.quote_id}>
                    {q.name}
                  </ComboboxItem>
                ))}
              </ComboboxList>
            </ComboboxContent>
          </Combobox>
        </div>
        <Button size="sm" onClick={handleAdd} disabled={!addQuoteId || busy}>
          Ajouter
        </Button>
      </div>

      {quotes.length === 0 && (
        <p className="text-sm text-muted-foreground">
          Aucun devis rattaché à ce projet.
        </p>
      )}

      {/* Quote rows */}
      {quotes.map((q) => {
        const isExpanded = expanded.has(q.quote_id);
        return (
          <div key={q.quote_id} className="rounded-lg border">
            {/* Main row */}
            <div className="flex items-center gap-3 p-3">
              <button
                type="button"
                onClick={() => toggle(q.quote_id)}
                className="text-muted-foreground hover:text-foreground"
              >
                {isExpanded ? (
                  <ChevronDownIcon className="size-4" />
                ) : (
                  <ChevronRightIcon className="size-4" />
                )}
              </button>
              <Link
                href={`/quote/${q.quote_id}`}
                className="flex-1 font-medium hover:underline"
              >
                {q.name}
              </Link>
              <Badge variant={QUOTE_STATE_VARIANTS[q.state] ?? "outline"}>
                {QUOTE_STATE_LABELS[q.state] ?? q.state}
              </Badge>
              <span className="text-xs text-muted-foreground">
                {q.schedules?.length ?? 0} éch. · {q.invoices?.length ?? 0}{" "}
                fact.
              </span>
              <Button
                variant="ghost"
                size="icon-xs"
                onClick={() => handleRemove(q.quote_id)}
                disabled={busy}
                title="Retirer du projet"
              >
                <UnlinkIcon className="size-3.5" />
              </Button>
            </div>

            {/* Expanded sub-tables */}
            {isExpanded && (
              <div className="border-t px-4 py-3 grid grid-cols-1 gap-4 md:grid-cols-2">
                <div>
                  <p className="mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                    Échéanciers
                  </p>
                  <ScheduleSubTable schedules={q.schedules ?? []} />
                </div>
                <div>
                  <p className="mb-2 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                    Factures
                  </p>
                  <InvoiceSubTable invoices={q.invoices ?? []} />
                </div>
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}
