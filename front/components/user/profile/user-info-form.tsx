"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  apiFetch,
  fieldErrorsFromBody,
  FieldErrors,
  toErrorProps,
} from "@/lib/api";
import { toast } from "sonner";

export type UserProfile = {
  user_id: string;
  email: string;
  phone: string;
  company: string;
  siren: string;
  vat: string;
};

type UserInfoFormProps = {
  user: UserProfile;
  onSaved?: (user: UserProfile) => void;
};

export default function UserInfoForm({ user, onSaved }: UserInfoFormProps) {
  const [phone, setPhone] = useState(user.phone ?? "");
  const [company, setCompany] = useState(user.company ?? "");
  const [siren, setSiren] = useState(user.siren ?? "");
  const [vat, setVat] = useState(user.vat ?? "");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setFieldErrors({});
    setSubmitting(true);
    try {
      const { ok, status, body } = await apiFetch("/api/users/me", {
        method: "PUT",
        body: JSON.stringify({ phone, company, siren, vat }),
      });
      if (ok && body.success) {
        toast.success("Informations mises à jour.");
        onSaved?.({ ...user, phone, company, siren, vat });
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

  return (
    <form
      className="grid max-w-3xl gap-4"
      onSubmit={handleSubmit}
      noValidate
    >
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="email">Email</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            value={user.email}
            readOnly
            aria-readonly
          />
          <FieldDescription>
            L&apos;adresse email se modifie depuis l&apos;onglet Connexion.
          </FieldDescription>
        </Field>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors.phone?.length}>
            <FieldLabel htmlFor="phone">Téléphone</FieldLabel>
            <Input
              id="phone"
              name="phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              aria-invalid={!!fieldErrors.phone?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors.phone)} />
          </Field>

          <Field data-invalid={!!fieldErrors.company?.length}>
            <FieldLabel htmlFor="company">Société</FieldLabel>
            <Input
              id="company"
              name="company"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
              aria-invalid={!!fieldErrors.company?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors.company)} />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors.siren?.length}>
            <FieldLabel htmlFor="siren">SIREN</FieldLabel>
            <Input
              id="siren"
              name="siren"
              value={siren}
              onChange={(e) => setSiren(e.target.value)}
              aria-invalid={!!fieldErrors.siren?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors.siren)} />
          </Field>

          <Field data-invalid={!!fieldErrors.vat?.length}>
            <FieldLabel htmlFor="vat">N° de TVA</FieldLabel>
            <Input
              id="vat"
              name="vat"
              value={vat}
              onChange={(e) => setVat(e.target.value)}
              aria-invalid={!!fieldErrors.vat?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors.vat)} />
          </Field>
        </div>
      </FieldGroup>

      <div className="flex justify-end">
        <Button type="submit" disabled={submitting}>
          {submitting ? "Enregistrement…" : "Enregistrer"}
        </Button>
      </div>
    </form>
  );
}
