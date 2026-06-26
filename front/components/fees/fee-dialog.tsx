"use client";

import { useEffect, useState } from "react";
import { useDialogSubmit } from "@/hooks/use-dialog-submit";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { toErrorProps } from "@/lib/api";
import { createFee, updateFee } from "@/lib/services/fees";
import { listAvailableTaxesForUser } from "@/lib/services/taxes";
import type { BackendFee, BackendTax, FeeCategory } from "@/types/backend";

type FeeDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  fee?: BackendFee | null;
  onSaved: () => void;
};

const FORM_ID = "fee-form";
const CATEGORIES: FeeCategory[] = ["fixed", "service"];

export default function FeeDialog({
  open,
  onOpenChange,
  fee,
  onSaved,
}: FeeDialogProps) {
  const t = useTranslations("fees.dialog");
  const tCategories = useTranslations("fees.categories");
  const tCommon = useTranslations("common");
  const isEdit = fee != null;

  const [category, setCategory] = useState<FeeCategory>(
    fee?.category ?? "fixed",
  );
  const [name, setName] = useState(fee?.name ?? "");
  const [unit, setUnit] = useState(fee?.unit ?? "");
  const [priceEuros, setPriceEuros] = useState(
    fee ? String(fee.unit_price / 100) : "",
  );
  const [taxId, setTaxId] = useState<number | null>(fee?.tax_id ?? null);
  const [taxes, setTaxes] = useState<BackendTax[]>([]);
  const { fieldErrors, submitting, submit } = useDialogSubmit(
    tCommon("errors.generic"),
  );

  // Load the user's available taxes so a fee can carry a default tax.
  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    listAvailableTaxesForUser(
      fee?.tax_id ? [fee.tax_id] : [],
    ).then(({ ok, body }) => {
      if (cancelled) return;
      setTaxes(ok && Array.isArray(body.taxes) ? (body.taxes as BackendTax[]) : []);
    });
    return () => {
      cancelled = true;
    };
  }, [open, fee?.tax_id]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    const payload = {
      category,
      name,
      unit,
      unit_price: Math.round(Number(priceEuros) * 100) || 0,
      tax_id: taxId ?? 0,
    };
    await submit({
      request: () =>
        isEdit ? updateFee(fee!.fee_id, payload) : createFee(payload),
      successMessage: isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  const selectedTax = taxId != null ? (taxes.find((t) => t.id === taxId) ?? null) : null;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="p-6 sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t("editTitle") : t("createTitle")}
          </DialogTitle>
        </DialogHeader>

        <form
          id={FORM_ID}
          className="grid gap-4"
          onSubmit={handleSubmit}
          noValidate
        >
          <FieldGroup>
            <Field data-invalid={!!fieldErrors.category?.length}>
              <FieldLabel htmlFor="fee_category">{t("categoryLabel")}</FieldLabel>
              <Select
                value={category}
                onValueChange={(v) => setCategory(v as FeeCategory)}
              >
                <SelectTrigger id="fee_category" name="category">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {CATEGORIES.map((c) => (
                    <SelectItem key={c} value={c}>
                      {tCategories(c)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <FieldError errors={toErrorProps(fieldErrors.category)} />
            </Field>

            <Field data-invalid={!!fieldErrors.name?.length}>
              <FieldLabel htmlFor="fee_name">{t("nameLabel")}</FieldLabel>
              <Input
                id="fee_name"
                name="name"
                placeholder={t("namePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
                aria-invalid={!!fieldErrors.name?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.name)} />
            </Field>

            <Field data-invalid={!!fieldErrors.unit_price?.length}>
              <FieldLabel htmlFor="fee_price">{t("priceLabel")}</FieldLabel>
              <Input
                id="fee_price"
                name="unit_price"
                type="text"
                inputMode="decimal"
                placeholder={t("pricePlaceholder")}
                value={priceEuros}
                onChange={(e) => setPriceEuros(e.target.value.replace(",", "."))}
                aria-invalid={!!fieldErrors.unit_price?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.unit_price)} />
            </Field>

            <Field>
              <FieldLabel htmlFor="fee_unit">{t("unitLabel")}</FieldLabel>
              <Input
                id="fee_unit"
                name="unit"
                placeholder={t("unitPlaceholder")}
                value={unit}
                onChange={(e) => setUnit(e.target.value)}
              />
            </Field>

            <Field>
              <FieldLabel htmlFor="fee_tax">{t("taxLabel")}</FieldLabel>
              <Combobox
                items={taxes}
                value={selectedTax}
                onValueChange={(item: BackendTax | null) =>
                  setTaxId(item ? item.id : null)
                }
                itemToStringLabel={(item: BackendTax) => item.name}
              >
                <ComboboxInput
                  id="fee_tax"
                  name="tax_id"
                  placeholder={t("taxPlaceholder")}
                />
                <ComboboxContent>
                  <ComboboxEmpty>{t("taxEmpty")}</ComboboxEmpty>
                  <ComboboxList>
                    {(tax: BackendTax) => (
                      <ComboboxItem key={tax.id} value={tax}>
                        {tax.name}
                      </ComboboxItem>
                    )}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
            </Field>
          </FieldGroup>
        </form>

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="outline">
              {tCommon("actions.cancel")}
            </Button>
          </DialogClose>
          <Button type="submit" form={FORM_ID} disabled={submitting}>
            {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
