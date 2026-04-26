"use client";

import { useEffect, useState } from "react";
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
import {
  apiFetch,
  fieldErrorsFromBody,
  FieldErrors,
  toErrorProps,
} from "@/lib/api";
import { toast } from "sonner";
import { XIcon } from "lucide-react";
import { type Country } from "@/components/address/address-form";
import { type CountryGroup } from "@/components/admin/types";

type CountryGroupDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  group?: CountryGroup | null;
  onSaved: () => void;
};

const FORM_ID = "country-group-form";

export default function CountryGroupDialog({
  open,
  onOpenChange,
  group,
  onSaved,
}: CountryGroupDialogProps) {
  const isEdit = group != null;
  const [name, setName] = useState(group?.name ?? "");
  const [members, setMembers] = useState<Country[]>(group?.countries ?? []);
  const [allCountries, setAllCountries] = useState<Country[]>([]);
  const [pendingAdd, setPendingAdd] = useState<Country | null>(null);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [memberMutating, setMemberMutating] = useState(false);

  useEffect(() => {
    if (!isEdit) return;
    let cancelled = false;
    apiFetch("/api/users/countries").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.countries)) {
        setAllCountries(body.countries as Country[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [isEdit]);

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
    setFieldErrors({});
    setSubmitting(true);
    const path = isEdit
      ? `/api/users/country-groups/${group!.id}`
      : "/api/users/country-groups";
    try {
      const { ok, status, body } = await apiFetch(path, {
        method: isEdit ? "PUT" : "POST",
        body: JSON.stringify({ name }),
      });
      if (ok && body.success) {
        toast.success(
          isEdit ? "Groupe mis à jour." : "Groupe ajouté.",
        );
        onSaved();
        onOpenChange(false);
        return;
      }
      if (status === 422 && Array.isArray(body.field_errors)) {
        setFieldErrors(fieldErrorsFromBody(body));
        return;
      }
      toast.error(body.message ?? "Une erreur est survenue.");
    } catch {
      toast.error("Une erreur est survenue.");
    } finally {
      setSubmitting(false);
    }
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
        toast.success("Pays ajouté au groupe.");
        setPendingAdd(null);
        await refreshMembers();
        onSaved();
      } else {
        toast.error(body.message ?? "Une erreur est survenue.");
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
        toast.success("Pays retiré du groupe.");
        await refreshMembers();
        onSaved();
      } else {
        toast.error(body.message ?? "Une erreur est survenue.");
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
            {isEdit ? "Modifier le groupe" : "Nouveau groupe"}
          </DialogTitle>
        </DialogHeader>

        <form id={FORM_ID} className="grid gap-4" onSubmit={handleSubmit} noValidate>
          <FieldGroup>
            <Field data-invalid={!!fieldErrors.name?.length}>
              <FieldLabel htmlFor="group_name">Nom</FieldLabel>
              <Input
                id="group_name"
                name="name"
                placeholder="Union européenne"
                value={name}
                onChange={(e) => setName(e.target.value)}
                aria-invalid={!!fieldErrors.name?.length}
              />
              <FieldError errors={toErrorProps(fieldErrors.name)} />
            </Field>
          </FieldGroup>
        </form>

        {isEdit && (
          <div className="grid gap-3 rounded-lg border p-4" data-slot="group-members">
            <h3 className="text-sm font-medium">Pays membres</h3>

            {members.length === 0 ? (
              <p className="text-muted-foreground text-sm">
                Aucun pays dans ce groupe.
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
                      <span className="text-muted-foreground"> — {country.name}</span>
                    </span>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      aria-label={`Retirer ${country.name}`}
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
                  Ajouter un pays
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
                    placeholder="Sélectionner un pays"
                  />
                  <ComboboxContent>
                    <ComboboxEmpty>Aucun pays disponible.</ComboboxEmpty>
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
                Ajouter
              </Button>
            </div>
          </div>
        )}

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="outline">
              Fermer
            </Button>
          </DialogClose>
          <Button type="submit" form={FORM_ID} disabled={submitting}>
            {submitting ? "Enregistrement…" : "Enregistrer"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
