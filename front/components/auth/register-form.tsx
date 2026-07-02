"use client";
import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import OAuthButtons from "@/components/auth/oauth-buttons";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Field,
  FieldDescription,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { useRouter } from "next/navigation";
import { toast } from "sonner";

import { CONSENT_VERSIONS } from "@/lib/consent-versions";

// Must stay in sync with backend/auth/actions/errors.go field validation codes.
const FIELD_VALIDATION_KEYS: Record<number, string> = {
  1: "required",
  2: "invalidFormat",
  3: "tooShort",
  4: "emailInUse",
};

type FieldErrors = Record<string, string[]>;

type FormEvent = React.FormEvent<HTMLFormElement>;

type RegisterField = "email" | "password";

function toErrorProps(messages: string[] | undefined) {
  return messages?.map((message) => ({ message }));
}

export default function RegisterForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const router = useRouter();
  const t = useTranslations("auth.register");
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const [cgvAccepted, setCgvAccepted] = useState(false);
  const [cgvError, setCgvError] = useState<string | null>(null);

  function toMessages(field: RegisterField, codes: number[]): string[] {
    return codes.map((code) => {
      if (field === "email" && code === 2) {
        return t("validation.emailInvalidFormat");
      }
      const key = FIELD_VALIDATION_KEYS[code];
      return key ? t(`validation.${key}`) : t("validation.unknown", { code });
    });
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setFieldErrors({});
    setConfirmError(null);
    setCgvError(null);

    const form = e.currentTarget;
    const data = new FormData(form);
    const email = data.get("email") as string;
    const password = data.get("password") as string;
    const confirmPassword = data.get("confirm-password") as string;

    if (password !== confirmPassword) {
      setConfirmError(t("confirmMismatch"));
      return;
    }

    if (!cgvAccepted) {
      setCgvError(t("cgvRequired"));
      return;
    }

    try {
      const response = await fetch("/api/auth/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Accept: "application/json",
        },
        body: JSON.stringify({ email, password }),
      });

      if (response.ok) {
        await fetch("/api/auth/consent", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ type: "cgv", version: CONSENT_VERSIONS.cgv }),
        });
        toast.success(t("successToast"));
        router.replace("/login");
        return;
      }

      const body = await response.json();

      if (response.status === 422 && Array.isArray(body.field_errors)) {
        const errors: FieldErrors = {};
        for (const entry of body.field_errors as {
          field: string;
          error_code: number[];
        }[]) {
          if (entry.field === "email" || entry.field === "password") {
            errors[entry.field] = toMessages(entry.field, entry.error_code);
            continue;
          }
          errors[entry.field] = entry.error_code.map((code) => {
            const key = FIELD_VALIDATION_KEYS[code];
            return key
              ? t(`validation.${key}`)
              : t("validation.unknown", { code });
          });
        }
        setFieldErrors(errors);
        return;
      }

      toast.error(t("failureToast"));
    } catch {
      toast.error(t("failureToast"));
    }
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} noValidate>
            <FieldGroup>
              <Field data-invalid={!!fieldErrors.email?.length}>
                <FieldLabel htmlFor="email">{t("emailLabel")}</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  name="email"
                  placeholder="m@example.com"
                  aria-invalid={!!fieldErrors.email?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.email)} />
                <FieldDescription>{t("emailHint")}</FieldDescription>
              </Field>
              <Field data-invalid={!!fieldErrors.password?.length}>
                <FieldLabel htmlFor="password">{t("passwordLabel")}</FieldLabel>
                <Input
                  id="password"
                  type="password"
                  name="password"
                  aria-invalid={!!fieldErrors.password?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.password)} />
                <FieldDescription>{t("passwordHint")}</FieldDescription>
              </Field>
              <Field data-invalid={!!confirmError}>
                <FieldLabel htmlFor="confirm-password">
                  {t("confirmPasswordLabel")}
                </FieldLabel>
                <Input
                  id="confirm-password"
                  type="password"
                  name="confirm-password"
                  aria-invalid={!!confirmError}
                />
                <FieldError
                  errors={
                    confirmError ? [{ message: confirmError }] : undefined
                  }
                />
                <FieldDescription>{t("confirmPasswordHint")}</FieldDescription>
              </Field>
              <Field data-invalid={!!cgvError}>
                <div className="flex items-start gap-2">
                  <Checkbox
                    id="cgv"
                    checked={cgvAccepted}
                    onCheckedChange={(v) => {
                      setCgvAccepted(v === true);
                      if (v) setCgvError(null);
                    }}
                    aria-invalid={!!cgvError}
                  />
                  <FieldLabel htmlFor="cgv" className="font-normal leading-snug">
                    {t.rich("cgvLabel", {
                      cgvLink: (chunks) => <a href="/cgv" target="_blank" className="underline">{chunks}</a>,
                      privacyLink: (chunks) => <a href="/politique-de-confidentialite" target="_blank" className="underline">{chunks}</a>,
                    })}
                  </FieldLabel>
                </div>
                <FieldError
                  errors={cgvError ? [{ message: cgvError }] : undefined}
                />
              </Field>
              <FieldGroup>
                <Field>
                  <Button type="submit">{t("submit")}</Button>
                  <OAuthButtons next="/quote" />
                  <FieldDescription className="px-6 text-center">
                    {t("loginPrompt")} <a href="/login">{t("loginLink")}</a>
                  </FieldDescription>
                </Field>
              </FieldGroup>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
