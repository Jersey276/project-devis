"use client";

import { useState } from "react";
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
import { apiFetch, toErrorProps } from "@/lib/api";
import { type Country } from "@/components/address/address-form";

type CountryDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  country?: Country | null;
  onSaved: () => void;
};

const FORM_ID = "country-form";

export default function CountryDialog({
  open,
  onOpenChange,
  country,
  onSaved,
}: CountryDialogProps) {
  const t = useTranslations("admin.countries.dialog");
  const tCommon = useTranslations("common");
  const isEdit = country != null;
  const [code, setCode] = useState(country?.code ?? "");
  const [name, setName] = useState(country?.name ?? "");
  const { fieldErrors, submitting, submit } = useDialogSubmit(tCommon("errors.generic"));

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    const path = isEdit ? `/api/users/countries/${country!.id}` : "/api/users/countries";
    await submit({
      request: () => apiFetch(path, { method: isEdit ? "PUT" : "POST", body: JSON.stringify({ code, name }) }),
      successMessage: isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="p-6">
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
            <Field data-invalid={!!fieldErrors.code?.length}>
              <FieldLabel htmlFor="country_code">{t("codeLabel")}</FieldLabel>
              <Input
                id="country_code"
                name="code"
                placeholder={t("codePlaceholder")}
                value={code}
                onChange={(e) => setCode(e.target.value)}
                aria-invalid={!!fieldErrors.code?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.code)} />
            </Field>

            <Field data-invalid={!!fieldErrors.name?.length}>
              <FieldLabel htmlFor="country_name">{t("nameLabel")}</FieldLabel>
              <Input
                id="country_name"
                name="name"
                placeholder={t("namePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
                aria-invalid={!!fieldErrors.name?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.name)} />
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
