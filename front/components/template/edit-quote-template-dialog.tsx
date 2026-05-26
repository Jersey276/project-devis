"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Loader2Icon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import QuoteStepItems, {
  type QuoteItemRow,
} from "@/components/quote/steps/quote-step-items";
import {
  createTemplateLine,
  deleteTemplateLine,
  getTemplate,
  listTemplateLines,
  updateTemplate,
  updateTemplateLine,
} from "@/lib/services/templates";
import type { BackendTax, BackendTemplate, BackendTemplateLine } from "@/types/backend";

type LocalItem = QuoteItemRow & { position: number };

type Props = {
  templateId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  availableTaxes: BackendTax[];
  taxById: Map<number, BackendTax>;
  onSaved: (updated: BackendTemplate) => void;
};

function rowFromLine(line: BackendTemplateLine): LocalItem {
  return {
    lineId: line.line_id,
    name: line.name,
    quantity: Number(line.quantity),
    unitPriceEuros: line.unit_price / 100,
    taxId: line.tax_id ?? null,
    position: line.position,
    saveStatus: "idle",
  };
}

let tempIdCounter = 0;
function newTempId(): string {
  return `temp-${++tempIdCounter}`;
}

export default function EditQuoteTemplateSheet({
  templateId,
  open,
  onOpenChange,
  availableTaxes,
  taxById,
  onSaved,
}: Props) {
  const t = useTranslations("templates.editQuote");
  const tCommon = useTranslations("common");

  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [adding, setAdding] = useState(false);

  const [templateName, setTemplateName] = useState("");
  const [originalName, setOriginalName] = useState("");
  const [items, setItems] = useState<LocalItem[]>([]);
  const [originalLineIds, setOriginalLineIds] = useState<Set<string>>(new Set());

  const itemsRef = useRef(items);
  itemsRef.current = items;

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setLoading(true);

    Promise.all([getTemplate(templateId), listTemplateLines(templateId)]).then(
      ([tplRes, linesRes]) => {
        if (cancelled) return;
        setLoading(false);

        if (!tplRes.ok || !tplRes.body.success) {
          toast.error(t("loadFailedToast"));
          onOpenChange(false);
          return;
        }
        const tpl = tplRes.body.template as BackendTemplate;
        setTemplateName(tpl.name);
        setOriginalName(tpl.name);

        const lines: LocalItem[] =
          linesRes.ok && Array.isArray(linesRes.body.lines)
            ? (linesRes.body.lines as BackendTemplateLine[])
                .sort((a, b) => a.position - b.position)
                .map(rowFromLine)
            : [];

        setItems(lines);
        setOriginalLineIds(new Set(lines.map((l) => l.lineId)));
      },
    );

    return () => {
      cancelled = true;
    };
  }, [open, templateId, t, onOpenChange]);

  const totals = useMemo(() => {
    const ht = items.reduce(
      (acc, item) => acc + item.quantity * item.unitPriceEuros,
      0,
    );
    return { ht, breakdown: [], ttc: ht };
  }, [items]);

  function setRow(lineId: string, patch: Partial<LocalItem>) {
    setItems((prev) =>
      prev.map((row) => (row.lineId === lineId ? { ...row, ...patch } : row)),
    );
  }

  function handleAddItem() {
    setAdding(true);
    const newItem: LocalItem = {
      lineId: newTempId(),
      name: "",
      quantity: 1,
      unitPriceEuros: 0,
      taxId: null,
      position: itemsRef.current.length,
      saveStatus: "idle",
    };
    setItems((prev) => [...prev, newItem]);
    setAdding(false);
  }

  function handleRemoveItem(lineId: string) {
    setItems((prev) => prev.filter((row) => row.lineId !== lineId));
  }

  async function handleSave() {
    setSaving(true);
    try {
      const trimmedName = templateName.trim();

      if (trimmedName !== originalName) {
        const tplRes = await updateTemplate(templateId, { name: trimmedName });
        if (!tplRes.ok || !tplRes.body.success) {
          toast.error(t("saveFailedToast"));
          return;
        }
      }

      const currentIds = new Set(items.map((i) => i.lineId));
      const deletedIds = [...originalLineIds].filter((id) => !currentIds.has(id));

      const results = await Promise.all([
        ...deletedIds.map((id) => deleteTemplateLine(templateId, id)),
        ...items.map((item, idx) => {
          const draft = {
            type: "simple",
            name: item.name,
            quantity: item.quantity,
            unitPriceEuros: item.unitPriceEuros,
            position: idx,
            taxId: item.taxId,
          };
          if (item.lineId.startsWith("temp-")) {
            return createTemplateLine(templateId, draft);
          }
          return updateTemplateLine(templateId, item.lineId, draft);
        }),
      ]);

      const failed = results.some((r) => !r.ok || !r.body.success);
      if (failed) {
        toast.error(t("saveFailedToast"));
        return;
      }

      const tplRes = await getTemplate(templateId);
      toast.success(t("saveSuccessToast"));
      if (tplRes.ok && tplRes.body.success) {
        onSaved(tplRes.body.template as BackendTemplate);
      }
      onOpenChange(false);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-3xl overflow-y-auto max-h-[90vh]">
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-10">
            <Loader2Icon className="text-muted-foreground size-6 animate-spin" />
          </div>
        ) : (
          <div className="flex flex-col gap-6">
            <div className="space-y-1.5">
              <Label>{t("templateNameLabel")}</Label>
              <Input
                value={templateName}
                onChange={(e) => setTemplateName(e.target.value)}
              />
            </div>

            <QuoteStepItems
              items={items}
              isReadonly={false}
              totals={totals}
              availableTaxes={availableTaxes}
              taxById={taxById}
              isAdding={adding}
              hideTotals
              onNameChange={(id, v) => setRow(id, { name: v })}
              onQuantityChange={(id, v) =>
                setRow(id, { quantity: Number.isFinite(v) ? v : 0 })
              }
              onUnitPriceChange={(id, v) =>
                setRow(id, { unitPriceEuros: Number.isFinite(v) ? v : 0 })
              }
              onTaxChange={(id, tid) => setRow(id, { taxId: tid })}
              onRemoveItem={handleRemoveItem}
              onAddItem={handleAddItem}
            />
          </div>
        )}

        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => onOpenChange(false)}
            disabled={saving}
          >
            {tCommon("actions.cancel")}
          </Button>
          <Button onClick={handleSave} disabled={loading || saving}>
            {saving ? tCommon("actions.saving") : tCommon("actions.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
