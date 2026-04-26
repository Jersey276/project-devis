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

type ConnectionFormProps = {
  email: string;
};

export default function ConnectionForm({ email }: ConnectionFormProps) {
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  function reset() {
    setOldPassword("");
    setNewPassword("");
    setConfirmPassword("");
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setFieldErrors({});
    setConfirmError(null);

    if (newPassword !== confirmPassword) {
      setConfirmError("Les mots de passe ne correspondent pas.");
      return;
    }

    setSubmitting(true);
    try {
      const { ok, status, body } = await apiFetch(
        "/api/auth/password/update",
        {
          method: "POST",
          body: JSON.stringify({
            email,
            old_password: oldPassword,
            new_password: newPassword,
          }),
        },
      );
      if (ok && body.success) {
        toast.success("Mot de passe mis à jour.");
        reset();
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
          <FieldLabel htmlFor="connection_email">Adresse mail</FieldLabel>
          <Input
            id="connection_email"
            name="email"
            type="email"
            value={email}
            readOnly
            aria-readonly
          />
          <FieldDescription>
            Le changement d&apos;adresse email n&apos;est pas encore disponible.
          </FieldDescription>
        </Field>

        <Field data-invalid={!!fieldErrors.old_password?.length}>
          <FieldLabel htmlFor="old_password">Mot de passe actuel</FieldLabel>
          <Input
            id="old_password"
            name="old_password"
            type="password"
            value={oldPassword}
            onChange={(e) => setOldPassword(e.target.value)}
            autoComplete="current-password"
            aria-invalid={!!fieldErrors.old_password?.length}
          />
          <FieldError errors={toErrorProps(fieldErrors.old_password)} />
        </Field>

        <Field data-invalid={!!fieldErrors.new_password?.length}>
          <FieldLabel htmlFor="new_password">Nouveau mot de passe</FieldLabel>
          <Input
            id="new_password"
            name="new_password"
            type="password"
            value={newPassword}
            onChange={(e) => setNewPassword(e.target.value)}
            autoComplete="new-password"
            aria-invalid={!!fieldErrors.new_password?.length}
          />
          <FieldError errors={toErrorProps(fieldErrors.new_password)} />
          <FieldDescription>8 caractères minimum.</FieldDescription>
        </Field>

        <Field data-invalid={!!confirmError}>
          <FieldLabel htmlFor="confirm_password">Confirmation</FieldLabel>
          <Input
            id="confirm_password"
            name="confirm_password"
            type="password"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            autoComplete="new-password"
            aria-invalid={!!confirmError}
          />
          <FieldError
            errors={confirmError ? [{ message: confirmError }] : undefined}
          />
        </Field>
      </FieldGroup>

      <div className="flex justify-end">
        <Button type="submit" disabled={submitting}>
          {submitting ? "Enregistrement…" : "Mettre à jour le mot de passe"}
        </Button>
      </div>
    </form>
  );
}
