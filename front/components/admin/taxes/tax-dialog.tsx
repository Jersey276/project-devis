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
import { Checkbox } from "@/components/ui/checkbox";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { apiFetch, toErrorProps } from "@/lib/api";
import { type CountryGroup, type Tax } from "@/components/admin/types";

type TaxDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  tax?: Tax | null;
  groups: CountryGroup[];
  onSaved: () => void;
};

const FORM_ID = "tax-form";

export default function TaxDialog({
  open,
  onOpenChange,
  tax,
  groups,
  onSaved,
}: TaxDialogProps) {
  const t = useTranslations("admin.taxes.dialog");
  const tCommon = useTranslations("common");
  const isEdit = tax != null;
  const [name, setName] = useState(tax?.name ?? "");
  const [rate, setRate] = useState(tax?.rate ?? "");
  const [groupId, setGroupId] = useState<number | null>(
    tax?.country_group_id ?? null,
  );
  const [isDefault, setIsDefault] = useState<boolean>(tax?.is_default ?? false);
  const { fieldErrors, submitting, submit } = useDialogSubmit(tCommon("errors.generic"));

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    const path = isEdit ? `/api/users/taxes/${tax!.id}` : "/api/users/taxes";
    const payload = isEdit
      ? { name, rate, is_default: isDefault }
      : { name, rate, country_group_id: groupId, is_default: isDefault };
    await submit({
      request: () => apiFetch(path, { method: isEdit ? "PUT" : "POST", body: JSON.stringify(payload) }),
      successMessage: isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  const selectedGroup =
    groupId != null ? (groups.find((g) => g.id === groupId) ?? null) : null;

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
            <Field data-invalid={!!fieldErrors.name?.length}>
              <FieldLabel htmlFor="tax_name">{t("nameLabel")}</FieldLabel>
              <Input
                id="tax_name"
                name="name"
                placeholder={t("namePlaceholder")}
                value={name}
                onChange={(e) => setName(e.target.value)}
                aria-invalid={!!fieldErrors.name?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.name)} />
            </Field>

            <Field data-invalid={!!fieldErrors.rate?.length}>
              <FieldLabel htmlFor="tax_rate">{t("rateLabel")}</FieldLabel>
              <Input
                id="tax_rate"
                name="rate"
                placeholder={t("ratePlaceholder")}
                value={rate}
                onChange={(e) => setRate(e.target.value)}
                aria-invalid={!!fieldErrors.rate?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.rate)} />
            </Field>

            <Field data-invalid={!!fieldErrors.country_group_id?.length}>
              <FieldLabel htmlFor="tax_country_group">
                {t("groupLabel")}
              </FieldLabel>
              <Combobox
                items={groups}
                value={selectedGroup}
                onValueChange={(item: CountryGroup | null) =>
                  setGroupId(item ? item.id : null)
                }
                itemToStringLabel={(item: CountryGroup) => item.name}
                disabled={isEdit}
              >
                <ComboboxInput
                  id="tax_country_group"
                  name="country_group_id"
                  placeholder={t("groupPlaceholder")}
                  disabled={isEdit}
                  aria-invalid={!!fieldErrors.country_group_id?.length}
                />
                <ComboboxContent>
                  <ComboboxEmpty>{t("groupEmpty")}</ComboboxEmpty>
                  <ComboboxList>
                    {(group: CountryGroup) => (
                      <ComboboxItem key={group.id} value={group}>
                        {group.name}
                      </ComboboxItem>
                    )}
                  </ComboboxList>
                </ComboboxContent>
              </Combobox>
              <FieldError errors={toErrorProps(fieldErrors.country_group_id)} />
            </Field>

            <Field>
              <label className="flex items-center gap-2 text-sm">
                <Checkbox
                  id="tax_is_default"
                  name="is_default"
                  checked={isDefault}
                  onCheckedChange={(v) => setIsDefault(v === true)}
                />
                Taxe par défaut du groupe
              </label>
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
