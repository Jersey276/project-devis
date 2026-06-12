"use client";

import { FormEvent, useMemo, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
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
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const CODE_INVALID_RESET_TOKEN = 1005;
const CODE_EXPIRED_RESET_TOKEN = 1006;
const CODE_WEAK_PASSWORD = 1007;

export default function ResetPasswordForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const t = useTranslations("auth.resetPassword");
  const params = useSearchParams();
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);
  const token = useMemo(() => params.get("token")?.trim() ?? "", [params]);

  if (!token) {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader>
            <CardTitle>{t("invalidLinkTitle")}</CardTitle>
            <CardDescription>{t("invalidLinkDescription")}</CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              type="button"
              onClick={() => router.replace("/forget-password")}
            >
              {t("requestNewLink")}
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setConfirmError(null);

    const form = e.currentTarget;
    const data = new FormData(form);
    const newPassword = String(data.get("new_password") ?? "");
    const confirmPassword = String(data.get("confirm_password") ?? "");

    if (newPassword !== confirmPassword) {
      setConfirmError(t("confirmMismatch"));
      return;
    }

    setLoading(true);
    const result = await apiFetch("/api/auth/password/confirm-reset", {
      method: "POST",
      body: JSON.stringify({ token, new_password: newPassword }),
    });
    setLoading(false);

    if (result.status === 429) {
      toast.error(t("rateLimitedToast"));
      return;
    }

    if (result.ok) {
      toast.success(t("successToast"));
      router.replace("/login");
      return;
    }

    if (result.body.code === CODE_INVALID_RESET_TOKEN) {
      toast.error(t("invalidTokenToast"));
      return;
    }
    if (result.body.code === CODE_EXPIRED_RESET_TOKEN) {
      toast.error(t("expiredTokenToast"));
      return;
    }
    if (result.body.code === CODE_WEAK_PASSWORD) {
      toast.error(t("weakPasswordToast"));
      return;
    }

    toast.error(t("failureToast"));
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
              <Field>
                <FieldLabel htmlFor="new_password">
                  {t("newPasswordLabel")}
                </FieldLabel>
                <Input
                  id="new_password"
                  type="password"
                  name="new_password"
                  required
                />
                <FieldDescription>{t("newPasswordHint")}</FieldDescription>
              </Field>
              <Field data-invalid={!!confirmError}>
                <FieldLabel htmlFor="confirm_password">
                  {t("confirmPasswordLabel")}
                </FieldLabel>
                <Input
                  id="confirm_password"
                  type="password"
                  name="confirm_password"
                  required
                  aria-invalid={!!confirmError}
                />
                <FieldError
                  errors={
                    confirmError ? [{ message: confirmError }] : undefined
                  }
                />
              </Field>
              <Field>
                <Button type="submit" disabled={loading}>
                  {loading ? t("submitLoading") : t("submit")}
                </Button>
                <FieldDescription className="text-center">
                  {t("backToLoginPrompt")}{" "}
                  <a href="/login">{t("backToLoginLink")}</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
