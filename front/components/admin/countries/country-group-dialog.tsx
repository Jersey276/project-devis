"use client";

import { useState } from "react";
import { useCountries } from "@/hooks/use-countries";
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
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { apiFetch, toErrorProps } from "@/lib/api";
import { toast } from "sonner";
import { XIcon } from "lucide-react";
import { type Country } from "@/components/address/address-form";
import { useDialogSubmit } from "@/hooks/use-dialog-submit";
import { type CountryGroup } from "@/components/admin/types";

type CountryGroupDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  group?: CountryGroup | null;
  onSaved: () => void;
  allCountries?: Country[];
};

const FORM_ID = "country-group-form";

export default function CountryGroupDialog({
  open,
  onOpenChange,
  group,
  onSaved,
  allCountries: allCountriesProp,
}: CountryGroupDialogProps) {
  const t = useTranslations("admin.countryGroups.dialog");
  const tCommon = useTranslations("common");
  const isEdit = group != null;
  const [name, setName] = useState(group?.name ?? "");
  const [members, setMembers] = useState<Country[]>(group?.countries ?? []);
  const fetchedCountries = useCountries(0, allCountriesProp !== undefined || !isEdit);
  const allCountries = allCountriesProp ?? fetchedCountries;
  const [pendingAdd, setPendingAdd] = useState<Country | null>(null);
  const { fieldErrors, submitting, submit } = useDialogSubmit(tCommon("errors.generic"));
  const [memberMutating, setMemberMutating] = useState(false);

  async function refreshMembers() {
    if (!group) return;
    const { ok, body } = await apiFetch(
      `/api/users/country-groups/${group.id}`,
    );
    if (ok && body.success) {
      const fresh = body.country_group as CountryGroup | undefined;
      setMembers(fresh?.countries ?? []);
    }
  }

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    const path = isEdit
      ? `/api/users/country-groups/${group!.id}`
      : "/api/users/country-groups";
    await submit({
      request: () => apiFetch(path, { method: isEdit ? "PUT" : "POST", body: JSON.stringify({ name }) }),
      successMessage: isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  async function handleAttach() {
    if (!group || !pendingAdd || memberMutating) return;
    setMemberMutating(true);
    try {
      const { ok, body } = await apiFetch(
        `/api/users/country-groups/${group.id}/countries/${pendingAdd.id}`,
        { method: "POST" },
      );
      if (ok && body.success) {
        toast.success(t("attachSuccessToast"));
        setPendingAdd(null);
        await refreshMembers();
        onSaved();
      } else {
        toast.error(body.message ?? tCommon("errors.generic"));
      }
    } finally {
      setMemberMutating(false);
    }
  }

  async function handleDetach(country: Country) {
    if (!group || memberMutating) return;
    setMemberMutating(true);
    try {
      const { ok, body } = await apiFetch(
        `/api/users/country-groups/${group.id}/countries/${country.id}`,
        { method: "DELETE" },
      );
      if (ok && body.success) {
        toast.success(t("detachSuccessToast"));
        await refreshMembers();
        onSaved();
      } else {
        toast.error(body.message ?? tCommon("errors.generic"));
      }
    } finally {
      setMemberMutating(false);
    }
  }

  const memberIds = new Set(members.map((m) => m.id));
  const attachable = allCountries.filter((c) => !memberIds.has(c.id));

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
              <FieldLabel htmlFor="group_name">{t("nameLabel")}</FieldLabel>
              <Input
                id="group_name"
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

        {isEdit && (
          <div
            className="grid gap-3 rounded-lg border p-4"
            data-slot="group-members"
          >
            <h3 className="text-sm font-medium">{t("membersTitle")}</h3>

            {members.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                {t("membersEmpty")}
              </p>
            ) : (
              <ul className="grid gap-1.5">
                {members.map((country) => (
                  <li
                    key={country.id}
                    data-slot="group-member"
                    className="flex items-center justify-between rounded-md border px-3 py-1.5 text-sm"
                  >
                    <span>
                      <span className="font-medium">{country.code}</span>
                      <span className="text-muted-foreground">
                        {" "}
                        — {country.name}
                      </span>
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      aria-label={t("removeAria", { name: country.name })}
                      data-slot="group-member-remove"
                      disabled={memberMutating}
                      onClick={() => handleDetach(country)}
                    >
                      <XIcon />
                    </Button>
                  </li>
                ))}
              </ul>
            )}

            <div className="flex items-end gap-2">
              <Field className="flex-1">
                <FieldLabel htmlFor="group_attach_country">
                  {t("attachLabel")}
                </FieldLabel>
                <Combobox
                  items={attachable}
                  value={pendingAdd}
                  onValueChange={(item: Country | null) => setPendingAdd(item)}
                  itemToStringLabel={(item: Country) => item.name}
                >
                  <ComboboxInput
                    id="group_attach_country"
                    name="attach_country_id"
                    placeholder={t("attachPlaceholder")}
                  />
                  <ComboboxContent>
                    <ComboboxEmpty>{t("attachEmpty")}</ComboboxEmpty>
                    <ComboboxList>
                      {(country: Country) => (
                        <ComboboxItem key={country.id} value={country}>
                          {country.name}
                        </ComboboxItem>
                      )}
                    </ComboboxList>
                  </ComboboxContent>
                </Combobox>
              </Field>
              <Button
                type="button"
                onClick={handleAttach}
                disabled={!pendingAdd || memberMutating}
              >
                {t("attachButton")}
              </Button>
            </div>
          </div>
        )}

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="outline">
              {tCommon("actions.close")}
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
