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
import { Checkbox } from "@/components/ui/checkbox";
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
  suspended: boolean;
  oss_enabled: boolean;
};

type UserInfoFormProps = {
  user: UserProfile;
  onSaved?: (user: UserProfile) => void;
  readOnly?: boolean;
};

export default function UserInfoForm({
  user,
  onSaved,
  readOnly = false,
}: UserInfoFormProps) {
  const t = useTranslations("profile.userInfo");
  const tCommon = useTranslations("common");
  const [phone, setPhone] = useState(user.phone ?? "");
  const [company, setCompany] = useState(user.company ?? "");
  const [siren, setSiren] = useState(user.siren ?? "");
  const [vat, setVat] = useState(user.vat ?? "");
  const [ossEnabled, setOssEnabled] = useState(user.oss_enabled ?? false);
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (readOnly) return;
    setFieldErrors({});
    setSubmitting(true);
    try {
      const { ok, status, body } = await apiFetch("/api/users/me", {
        method: "PUT",
        body: JSON.stringify({
          phone,
          company,
          siren,
          vat,
          oss_enabled: ossEnabled,
        }),
      });
      if (ok && body.success) {
        toast.success(t("successToast"));
        onSaved?.({ ...user, phone, company, siren, vat, oss_enabled: ossEnabled });
        return;
      }
      if (status === 422 && Array.isArray(body.field_errors)) {
        setFieldErrors(fieldErrorsFromBody(body));
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
    <form className="grid max-w-3xl gap-4" onSubmit={handleSubmit} noValidate>
      <FieldGroup>
        {readOnly ? (
          <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
            {t("suspendedNotice")}
          </p>
        ) : null}

        <Field>
          <FieldLabel htmlFor="email">{t("emailLabel")}</FieldLabel>
          <Input
            id="email"
            name="email"
            type="email"
            value={user.email}
            readOnly
            aria-readonly
          />
          <FieldDescription>{t("emailHint")}</FieldDescription>
        </Field>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors.phone?.length}>
            <FieldLabel htmlFor="phone">{t("phoneLabel")}</FieldLabel>
            <Input
              id="phone"
              name="phone"
              type="tel"
              value={phone}
              onChange={(e) => setPhone(e.target.value)}
              aria-invalid={!!fieldErrors.phone?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.phone)} />
          </Field>

          <Field data-invalid={!!fieldErrors.company?.length}>
            <FieldLabel htmlFor="company">{t("companyLabel")}</FieldLabel>
            <Input
              id="company"
              name="company"
              value={company}
              onChange={(e) => setCompany(e.target.value)}
              aria-invalid={!!fieldErrors.company?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.company)} />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors.siren?.length}>
            <FieldLabel htmlFor="siren">{t("sirenLabel")}</FieldLabel>
            <Input
              id="siren"
              name="siren"
              value={siren}
              onChange={(e) => setSiren(e.target.value)}
              aria-invalid={!!fieldErrors.siren?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.siren)} />
          </Field>

          <Field data-invalid={!!fieldErrors.vat?.length}>
            <FieldLabel htmlFor="vat">{t("vatLabel")}</FieldLabel>
            <Input
              id="vat"
              name="vat"
              value={vat}
              onChange={(e) => setVat(e.target.value)}
              aria-invalid={!!fieldErrors.vat?.length}
              disabled={readOnly}
            />
            <FieldError errors={toErrorProps(fieldErrors.vat)} />
          </Field>
        </div>

        <Field orientation="horizontal">
          <Checkbox
            id="oss_enabled"
            checked={ossEnabled}
            onCheckedChange={(checked) => setOssEnabled(checked === true)}
            disabled={readOnly}
          />
          <div className="grid gap-1">
            <FieldLabel htmlFor="oss_enabled">{t("ossEnabledLabel")}</FieldLabel>
            <FieldDescription>{t("ossEnabledHint")}</FieldDescription>
          </div>
        </Field>
      </FieldGroup>

      <div className="flex justify-end">
        <Button type="submit" disabled={submitting || readOnly}>
          {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
        </Button>
      </div>
    </form>
  );
}
