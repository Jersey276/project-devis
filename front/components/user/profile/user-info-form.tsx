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
  first_name: string;
  last_name: string;
  phone: string;
  company: string;
  siren: string;
  siret: string;
  vat: string;
  suspended: boolean;
  oss_enabled: boolean;
  iban: string;
  bic: string;
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
  const [firstName, setFirstName] = useState(user.first_name ?? "");
  const [lastName, setLastName] = useState(user.last_name ?? "");
  const [phone, setPhone] = useState(user.phone ?? "");
  const [company, setCompany] = useState(user.company ?? "");
  const [siren, setSiren] = useState(user.siren ?? "");
  const [siret, setSiret] = useState(user.siret ?? "");
  const [vat, setVat] = useState(user.vat ?? "");
  const [ossEnabled, setOssEnabled] = useState(user.oss_enabled ?? false);
  const [iban, setIban] = useState(user.iban ?? "");
  const [bic, setBic] = useState(user.bic ?? "");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  // Mirrors backend sqlutil.ValidateSIRET: empty is allowed; otherwise 14 digits
  // and, when a SIREN is set, the SIRET must start with it. Catches the bad value
  // before the round-trip; the server still validates authoritatively.
  function validateSiret(): string | null {
    const s = siret.replace(/\s/g, "");
    if (s === "") return null;
    if (!/^\d{14}$/.test(s)) return t("siretInvalidLength");
    const sn = siren.replace(/\s/g, "");
    if (sn !== "" && !s.startsWith(sn)) return t("siretSirenMismatch");
    return null;
  }

  async function handleSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault();
    if (readOnly) return;
    setFieldErrors({});
    const siretError = validateSiret();
    if (siretError) {
      setFieldErrors({ siret: [siretError] });
      return;
    }
    setSubmitting(true);
    try {
      const { ok, body } = await apiFetch("/api/users/me", {
        method: "PUT",
        body: JSON.stringify({
          first_name: firstName,
          last_name: lastName,
          phone,
          company,
          siren,
          siret,
          vat,
          oss_enabled: ossEnabled,
          iban,
          bic,
        }),
      });
      if (ok && body.success) {
        toast.success(t("successToast"));
        onSaved?.({ ...user, first_name: firstName, last_name: lastName, phone, company, siren, siret, vat, oss_enabled: ossEnabled, iban, bic });
        return;
      }
      const parsed = fieldErrorsFromBody(body);
      if (Object.keys(parsed).length > 0) {
        setFieldErrors(parsed);
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
    <form className="grid max-w-3xl gap-8" onSubmit={handleSubmit} noValidate>
      {readOnly ? (
        <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
          {t("suspendedNotice")}
        </p>
      ) : null}

      {/* Informations de base */}
      <section className="grid gap-4">
        <h3 className="text-sm font-semibold">{t("sectionBasic")}</h3>
        <FieldGroup>
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
            <Field data-invalid={!!fieldErrors.first_name?.length}>
              <FieldLabel htmlFor="first_name">{t("firstNameLabel")}</FieldLabel>
              <Input
                id="first_name"
                name="first_name"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                aria-invalid={!!fieldErrors.first_name?.length}
                disabled={readOnly}
              />
              <FieldError errors={toErrorProps(fieldErrors.first_name)} />
            </Field>

            <Field data-invalid={!!fieldErrors.last_name?.length}>
              <FieldLabel htmlFor="last_name">{t("lastNameLabel")}</FieldLabel>
              <Input
                id="last_name"
                name="last_name"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                aria-invalid={!!fieldErrors.last_name?.length}
                disabled={readOnly}
              />
              <FieldError errors={toErrorProps(fieldErrors.last_name)} />
            </Field>
          </div>

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
        </FieldGroup>
      </section>

      {/* Informations de l'entreprise */}
      <section className="grid gap-4">
        <h3 className="text-sm font-semibold">{t("sectionCompany")}</h3>
        <FieldGroup>
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

            <Field data-invalid={!!fieldErrors.siret?.length}>
              <FieldLabel htmlFor="siret">{t("siretLabel")}</FieldLabel>
              <Input
                id="siret"
                name="siret"
                value={siret}
                onChange={(e) => setSiret(e.target.value)}
                aria-invalid={!!fieldErrors.siret?.length}
                disabled={readOnly}
              />
              <FieldError errors={toErrorProps(fieldErrors.siret)} />
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
      </section>

      {/* Informations bancaires */}
      <section className="grid gap-4">
        <h3 className="text-sm font-semibold">{t("sectionBanking")}</h3>
        <FieldGroup>
          <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
            <Field data-invalid={!!fieldErrors.iban?.length}>
              <FieldLabel htmlFor="iban">{t("ibanLabel")}</FieldLabel>
              <Input
                id="iban"
                name="iban"
                value={iban}
                onChange={(e) => setIban(e.target.value)}
                aria-invalid={!!fieldErrors.iban?.length}
                disabled={readOnly}
              />
              <FieldDescription>{t("ibanHint")}</FieldDescription>
              <FieldError errors={toErrorProps(fieldErrors.iban)} />
            </Field>

            <Field data-invalid={!!fieldErrors.bic?.length}>
              <FieldLabel htmlFor="bic">{t("bicLabel")}</FieldLabel>
              <Input
                id="bic"
                name="bic"
                value={bic}
                onChange={(e) => setBic(e.target.value)}
                aria-invalid={!!fieldErrors.bic?.length}
                disabled={readOnly}
              />
              <FieldError errors={toErrorProps(fieldErrors.bic)} />
            </Field>
          </div>
        </FieldGroup>
      </section>

      <div className="flex justify-end">
        <Button type="submit" disabled={submitting || readOnly}>
          {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
        </Button>
      </div>
    </form>
  );
}
