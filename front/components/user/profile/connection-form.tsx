"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
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
import OAuthAccounts from "@/components/user/profile/oauth-accounts";

// ─── Email section ────────────────────────────────────────────────────────────

function EmailSection({ currentEmail, readOnly }: { currentEmail: string; readOnly: boolean }) {
  const t = useTranslations("profile.connection");
  const tCommon = useTranslations("common");
  const [newEmail, setNewEmail] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [sent, setSent] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (readOnly) return;
    setSubmitting(true);
    try {
      const { ok, body } = await apiFetch("/api/auth/email/request-change", {
        method: "POST",
        body: JSON.stringify({ new_email: newEmail }),
      });
      if (ok && body.success) {
        setSent(true);
        toast.success(t("emailChangeSentToast"));
        return;
      }
      toast.error(body.message ?? tCommon("errors.generic"));
    } catch {
      toast.error(tCommon("errors.generic"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="grid max-w-3xl gap-4">
      <div className="space-y-1">
        <h3 className="text-sm font-semibold">{t("emailSectionTitle")}</h3>
        <p className="text-muted-foreground text-sm">{t("emailSectionHint")}</p>
      </div>

      {readOnly ? (
        <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
          {t("suspendedNotice")}
        </p>
      ) : sent ? (
        <p className="rounded-md border border-dashed px-3 py-2 text-sm text-green-700">
          {t("emailChangeSentNotice", { email: newEmail })}
        </p>
      ) : (
        <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="current_email">{t("currentEmailLabel")}</FieldLabel>
              <Input
                id="current_email"
                name="current_email"
                type="email"
                value={currentEmail}
                readOnly
                aria-readonly
              />
            </Field>

            <Field>
              <FieldLabel htmlFor="new_email">{t("newEmailLabel")}</FieldLabel>
              <Input
                id="new_email"
                name="new_email"
                type="email"
                value={newEmail}
                onChange={(e) => setNewEmail(e.target.value)}
                autoComplete="email"
                disabled={readOnly}
              />
              <FieldDescription>{t("newEmailHint")}</FieldDescription>
            </Field>
          </FieldGroup>

          <div className="flex justify-end">
            <Button type="submit" disabled={submitting || !newEmail}>
              {submitting ? tCommon("actions.saving") : t("emailChangeSubmit")}
            </Button>
          </div>
        </form>
      )}
    </section>
  );
}

// ─── Password section ─────────────────────────────────────────────────────────

function PasswordSection({ readOnly }: { readOnly: boolean }) {
  const t = useTranslations("profile.connection");
  const tCommon = useTranslations("common");
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
    if (readOnly) return;
    setFieldErrors({});
    setConfirmError(null);

    if (newPassword !== confirmPassword) {
      setConfirmError(t("confirmMismatch"));
      return;
    }

    setSubmitting(true);
    try {
      const { ok, status, body } = await apiFetch("/api/auth/password/update", {
        method: "POST",
        body: JSON.stringify({
          old_password: oldPassword,
          new_password: newPassword,
        }),
      });
      if (ok && body.success) {
        toast.success(t("successToast"));
        reset();
        return;
      }
      if (status === 422) {
        const parsed = fieldErrorsFromBody(body);
        if (Object.keys(parsed).length > 0) {
          setFieldErrors(parsed);
          return;
        }
      }
      toast.error(body.message ?? tCommon("errors.generic"));
    } catch {
      toast.error(tCommon("errors.generic"));
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <section className="grid max-w-3xl gap-4">
      <div className="space-y-1">
        <h3 className="text-sm font-semibold">{t("passwordSectionTitle")}</h3>
      </div>

      <form className="grid gap-4" onSubmit={handleSubmit} noValidate>
        <FieldGroup>
          {readOnly ? (
            <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
              {t("suspendedNotice")}
            </p>
          ) : null}

          <Field data-invalid={!!fieldErrors.old_password?.length}>
            <FieldLabel htmlFor="old_password">{t("oldPasswordLabel")}</FieldLabel>
            <Input
              id="old_password"
              name="old_password"
              type="password"
              value={oldPassword}
              onChange={(e) => setOldPassword(e.target.value)}
              autoComplete="current-password"
              aria-invalid={!!fieldErrors.old_password?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.old_password)} />
          </Field>

          <Field data-invalid={!!fieldErrors.new_password?.length}>
            <FieldLabel htmlFor="new_password">{t("newPasswordLabel")}</FieldLabel>
            <Input
              id="new_password"
              name="new_password"
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
              aria-invalid={!!fieldErrors.new_password?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.new_password)} />
            <FieldDescription>{t("newPasswordHint")}</FieldDescription>
          </Field>

          <Field data-invalid={!!confirmError}>
            <FieldLabel htmlFor="confirm_password">{t("confirmPasswordLabel")}</FieldLabel>
            <Input
              id="confirm_password"
              name="confirm_password"
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
              aria-invalid={!!confirmError}
              disabled={readOnly}
            />
            <FieldError
              errors={confirmError ? [{ message: confirmError }] : undefined}
            />
          </Field>
        </FieldGroup>

        <div className="flex justify-end">
          <Button type="submit" disabled={submitting || readOnly}>
            {submitting ? tCommon("actions.saving") : t("submit")}
          </Button>
        </div>
      </form>
    </section>
  );
}

// ─── Root ─────────────────────────────────────────────────────────────────────

type ConnectionFormProps = {
  email: string;
  readOnly?: boolean;
};

export default function ConnectionForm({
  email,
  readOnly = false,
}: ConnectionFormProps) {
  return (
    <div className="grid gap-10">
      <EmailSection currentEmail={email} readOnly={readOnly} />
      <PasswordSection readOnly={readOnly} />
      <OAuthAccounts readOnly={readOnly} />
    </div>
  );
}
