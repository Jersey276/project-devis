"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { setInvoiceLifecycleStatus } from "@/lib/services/invoices";
import { allowedNextLifecycleStatuses } from "@/lib/invoice-lifecycle";
import type { BackendInvoiceLifecycleStatus } from "@/types/backend";

type RealLifecycleStatus = Exclude<BackendInvoiceLifecycleStatus, "NONE">;

type Props = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  invoiceId: string;
  current: BackendInvoiceLifecycleStatus;
  onApplied: () => void;
};

export default function AdvanceLifecycleDialog({
  open,
  onOpenChange,
  invoiceId,
  current,
  onApplied,
}: Props) {
  const t = useTranslations("invoice.lifecycle");
  const options = allowedNextLifecycleStatuses(current);
  const [target, setTarget] = useState<RealLifecycleStatus | "">("");
  const [note, setNote] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleOpenChange(next: boolean) {
    if (!next) {
      setTarget("");
      setNote("");
      setError(null);
    }
    onOpenChange(next);
  }

  async function onConfirm() {
    if (!target) return;
    setBusy(true);
    const { ok, body } = await setInvoiceLifecycleStatus(invoiceId, target, note);
    setBusy(false);
    if (!ok || !body.success) {
      setError((body.message as string) ?? t("actionError"));
      return;
    }
    handleOpenChange(false);
    onApplied();
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t("advance")}</DialogTitle>
        </DialogHeader>
        <div className="space-y-4">
          {error ? <p className="text-sm text-destructive">{error}</p> : null}
          <Select
            value={target}
            onValueChange={(v) => setTarget(v as RealLifecycleStatus)}
          >
            <SelectTrigger>
              <SelectValue placeholder={t("targetPlaceholder")} />
            </SelectTrigger>
            <SelectContent>
              {options.map((s) => (
                <SelectItem key={s} value={s}>
                  {t(`status.${s}`)}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Textarea
            placeholder={t("note")}
            value={note}
            onChange={(e) => setNote(e.target.value)}
          />
        </div>
        <DialogFooter>
          <Button
            type="button"
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={busy}
          >
            {t("cancel")}
          </Button>
          <Button type="button" onClick={onConfirm} disabled={busy || !target}>
            {t("confirm")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
