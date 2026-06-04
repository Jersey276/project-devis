"use client";

import { useEffect, useState } from "react";
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
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import {
  getTemplate,
  listTemplateLines,
  updateTemplate,
  updateTemplateLine,
} from "@/lib/services/templates";
import type {
  BackendTax,
  BackendTemplate,
  BackendTemplateLine,
  QuoteLineData,
} from "@/types/backend";

type Props = {
  templateId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  availableTaxes: BackendTax[];
  taxById: Map<number, BackendTax>;
  onSaved: (updated: BackendTemplate) => void;
};

export default function EditLineTemplateSheet({
  templateId,
  open,
  onOpenChange,
  availableTaxes,
  taxById,
  onSaved,
}: Props) {
  const t = useTranslations("templates.editLine");
  const tCommon = useTranslations("common");

  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const [templateName, setTemplateName] = useState("");
  const [lineId, setLineId] = useState<string | null>(null);
  const [lineName, setLineName] = useState("");
  const [quantity, setQuantity] = useState(1);
  const [unitPriceEuros, setUnitPriceEuros] = useState(0);
  const [taxId, setTaxId] = useState<number | null>(null);
  const [lineData, setLineData] = useState<QuoteLineData>({ kind: "line" });

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

        if (
          linesRes.ok &&
          Array.isArray(linesRes.body.lines) &&
          linesRes.body.lines.length > 0
        ) {
          const line = (linesRes.body.lines as BackendTemplateLine[])[0];
          setLineId(line.line_id);
          setLineName(line.name);
          setQuantity(Number(line.quantity));
          setUnitPriceEuros(line.unit_price / 100);
          setTaxId(line.tax_id ?? null);
          setLineData(line.data);
        } else {
          setLineId(null);
          setLineName("");
          setQuantity(1);
          setUnitPriceEuros(0);
          setTaxId(null);
          setLineData({ kind: "line" });
        }
      },
    );

    return () => {
      cancelled = true;
    };
  }, [open, templateId, t, onOpenChange]);

  const selectableTaxes = availableTaxes.filter((tax) => !tax.superseded_at);
  const taxesDisabled = selectableTaxes.length === 0;
  const selectedTax = taxId != null ? (taxById.get(taxId) ?? null) : null;

  function taxLabel(tax: BackendTax): string {
    return `${tax.name} (${tax.rate}%)`;
  }

  async function handleSave() {
    setSaving(true);
    try {
      const tplRes = await updateTemplate(templateId, {
        name: templateName.trim(),
      });
      if (!tplRes.ok || !tplRes.body.success) {
        toast.error(t("saveFailedToast"));
        return;
      }

      if (lineId) {
        const lineRes = await updateTemplateLine(templateId, lineId, {
          type: "simple",
          name: lineName,
          quantity,
          unitPriceEuros,
          position: 0,
          taxId,
          data: lineData,
        });
        if (!lineRes.ok || !lineRes.body.success) {
          toast.error(t("saveFailedToast"));
          return;
        }
      }

      const refreshed = await getTemplate(templateId);
      if (!refreshed.ok || !refreshed.body.success) {
        toast.error(t("saveFailedToast"));
        return;
      }

      toast.success(t("saveSuccessToast"));
      onSaved(refreshed.body.template as BackendTemplate);
      onOpenChange(false);
    } finally {
      setSaving(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg p-6">
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
        </DialogHeader>

        {loading ? (
          <div className="flex items-center justify-center py-10">
            <Loader2Icon className="text-muted-foreground size-6 animate-spin" />
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            <div className="space-y-1.5">
              <Label>{t("templateNameLabel")}</Label>
              <Input
                value={templateName}
                onChange={(e) => setTemplateName(e.target.value)}
              />
            </div>

            <div className="space-y-1.5">
              <Label>{t("lineNameLabel")}</Label>
              <Input
                value={lineName}
                onChange={(e) => setLineName(e.target.value)}
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-1.5">
                <Label>{t("quantityLabel")}</Label>
                <Input
                  type="number"
                  min={0}
                  value={quantity}
                  onChange={(e) => setQuantity(Number(e.target.value))}
                />
              </div>

              <div className="space-y-1.5">
                <Label>{t("unitPriceLabel")}</Label>
                <Input
                  type="number"
                  min={0}
                  step="0.01"
                  value={unitPriceEuros}
                  onChange={(e) => setUnitPriceEuros(Number(e.target.value))}
                />
              </div>
            </div>

            <div className="space-y-1.5">
              <Label>{t("taxLabel")}</Label>
              <Combobox
                items={availableTaxes}
                value={selectedTax}
                onValueChange={(tax: BackendTax | null) =>
                  setTaxId(tax ? tax.id : null)
                }
                itemToStringLabel={taxLabel}
                disabled={taxesDisabled}
              >
                <ComboboxInput
                  placeholder={taxesDisabled ? "—" : undefined}
                  disabled={taxesDisabled}
                />
                <ComboboxContent>
                  <ComboboxEmpty>—</ComboboxEmpty>
                  <ComboboxList>
                    {(tax: BackendTax) => (
                      <ComboboxItem
                        key={tax.id}
                        value={tax}
                        disabled={!!tax.superseded_at}
                      >
                        {taxLabel(tax)}
                      </ComboboxItem>
                    )}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
            </div>
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
