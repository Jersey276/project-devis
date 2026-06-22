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
import { Label } from "@/components/ui/label";
import { FieldError } from "@/components/ui/field";
import { createInvoiceFromSchedule } from "@/lib/services/invoices";
import { formatEurosFromCents } from "@/lib/utils";
import type { BackendScheduleColumnTotal } from "@/types/backend";

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  scheduleId: string;
  durationMonths: number;
  columnTotals: BackendScheduleColumnTotal[];
  onCreated?: (invoiceId: string) => void;
};

export default function GenerateInvoiceFromScheduleDialog({
  open,
  onOpenChange,
  scheduleId,
  durationMonths,
  columnTotals,
  onCreated,
}: Props) {
  const t = useTranslations("invoice.generateFromSchedule");
  const router = useRouter();
  const [selected, setSelected] = useState<Set<number>>(new Set());
  const [issueNow, setIssueNow] = useState(true);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState("");

  const totalsByMonth = useMemo(() => {
    const m = new Map<number, number>();
    for (const c of columnTotals) m.set(c.month_index, c.amount_cents);
    return m;
  }, [columnTotals]);

  const months = useMemo(
    () => Array.from({ length: durationMonths }, (_, i) => i + 1),
    [durationMonths],
  );

  const selectedTotal = useMemo(() => {
    let sum = 0;
    for (const month of selected) sum += totalsByMonth.get(month) ?? 0;
    return sum;
  }, [selected, totalsByMonth]);

  function toggle(month: number) {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(month)) next.delete(month);
      else next.add(month);
      return next;
    });
  }

  async function onSubmit() {
    setError("");
    setSubmitting(true);
    const monthIndexes = Array.from(selected).sort((a, b) => a - b);
    const { ok, body } = await createInvoiceFromSchedule({
      scheduleId,
      monthIndexes,
      issueNow,
    });
    setSubmitting(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("error"));
      return;
    }
    const invoiceId = (body.invoice_id as string) ?? "";
    handleOpenChange(false);
    onCreated?.(invoiceId);
    if (invoiceId) router.push(`/invoice/${invoiceId}`);
  }

  function handleOpenChange(next: boolean) {
    if (!next) {
      setSelected(new Set());
      setError("");
      setIssueNow(true);
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
          <p className="text-sm text-muted-foreground">{t("description")}</p>

          <div className="grid grid-cols-2 gap-2">
            {months.map((month) => {
              const id = `inv-month-${month}`;
              return (
                <label
                  key={month}
                  htmlFor={id}
                  className="flex items-center justify-between gap-2 rounded border p-2 text-sm"
                >
                  <span className="flex items-center gap-2">
                    <Checkbox
                      id={id}
                      checked={selected.has(month)}
                      onCheckedChange={() => toggle(month)}
                    />
                    {t("month", { index: month })}
                  </span>
                  <span className="tabular-nums text-muted-foreground">
                    {formatEurosFromCents(totalsByMonth.get(month) ?? 0)}
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

          <label htmlFor="inv-issue-now" className="flex items-center gap-2 text-sm">
            <Checkbox
              id="inv-issue-now"
              checked={issueNow}
              onCheckedChange={(v) => setIssueNow(v === true)}
            />
            <Label htmlFor="inv-issue-now" className="cursor-pointer font-normal">
              {t("issueNow")}
            </Label>
          </label>

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
