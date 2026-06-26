"use client";

import { useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { FieldError } from "@/components/ui/field";
import { createCreditNote } from "@/lib/services/invoices";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendInvoiceDetails } from "@/types/backend";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  invoice: BackendInvoiceDetails;
  onCreated?: (creditNoteId: string) => void;
};

export default function CreateCreditNoteDialog({
  open,
  onOpenChange,
  invoice,
  onCreated,
}: Props) {
  const t = useTranslations("creditNote.create");
  const router = useRouter();
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [reason, setReason] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const creditedSet = useMemo(
    () => new Set(invoice.credited_positions ?? []),
    [invoice.credited_positions],
  );

  // Lines are indexed by their array position, which matches the backend's
  // invoice_line_snapshots.position (lines are returned ordered by position).
  const remaining = useMemo(
    () => invoice.lines.map((_, i) => i).filter((i) => !creditedSet.has(i)),
    [invoice.lines, creditedSet],
  );

  const selectedTotal = useMemo(() => {
    let sum = 0;
    for (const i of selected) sum += invoice.lines[i]?.line_ht_cents ?? 0;
    return sum;
  }, [selected, invoice.lines]);

  function toggle(i: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(i)) next.delete(i);
      else next.add(i);
      return next;
    });
  }

  function selectAllRemaining() {
    setSelected(new Set(remaining));
  }

  async function onSubmit() {
    setError("");
    setSubmitting(true);
    const positions = Array.from(selected).sort((a, b) => a - b);
    const { ok, body } = await createCreditNote(invoice.invoice_id, {
      positions,
      reason,
    });
    setSubmitting(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("error"));
      return;
    }
    const creditNoteId = (body.credit_note_id as string) ?? "";
    handleOpenChange(false);
    onCreated?.(creditNoteId);
    if (creditNoteId) router.push(`/credit-note/${creditNoteId}`);
  }

  function handleOpenChange(next: boolean) {
    if (!next) {
      setSelected(new Set());
      setReason("");
      setError("");
    }
    onOpenChange(next);
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
        </DialogHeader>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <p className="text-sm text-muted-foreground">{t("selectLines")}</p>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={selectAllRemaining}
              disabled={remaining.length === 0}
            >
              {t("totalCredit")}
            </Button>
          </div>

          <div className="space-y-1">
            {invoice.lines.map((line, i) => {
              const credited = creditedSet.has(i);
              const id = `cn-line-${i}`;
              return (
                <label
                  key={i}
                  htmlFor={id}
                  className="flex items-center justify-between gap-2 rounded border p-2 text-sm aria-disabled:opacity-50"
                  aria-disabled={credited}
                >
                  <span className="flex items-center gap-2">
                    <Checkbox
                      id={id}
                      checked={credited || selected.has(i)}
                      disabled={credited}
                      onCheckedChange={() => toggle(i)}
                    />
                    <span>
                      {line.name}
                      {credited ? (
                        <span className="ml-2 text-xs text-muted-foreground">
                          ({t("alreadyCredited")})
                        </span>
                      ) : null}
                    </span>
                  </span>
                  <span className="tabular-nums text-muted-foreground">
                    {formatEurosFromCents(line.line_ht_cents)}
                  </span>
                </label>
              );
            })}
          </div>

          <div className="flex items-center justify-between border-t pt-2 text-sm font-medium">
            <span>{t("selectedTotal")}</span>
            <span className="tabular-nums">
              {formatEurosFromCents(selectedTotal)}
            </span>
          </div>

          <div className="space-y-1">
            <Label htmlFor="cn-reason" className="text-sm font-normal">
              {t("reason")}
            </Label>
            <Input
              id="cn-reason"
              value={reason}
              onChange={(e) => setReason(e.target.value)}
            />
          </div>

          <FieldError>{error}</FieldError>
        </div>

        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
          >
            {t("cancel")}
          </Button>
          <Button
            type="button"
            onClick={onSubmit}
            disabled={submitting || selected.size === 0}
          >
            {submitting ? t("submitting") : t("submit")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
