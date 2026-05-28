"use client";

import { FormEvent, useState } from "react";
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
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

type FormDataShape = {
  email: string;
};

export default function ForgetPasswordForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const t = useTranslations("auth.forgetPassword");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoading(true);

    const form = e.currentTarget;
    const data = new FormData(form);
    const payload: FormDataShape = {
      email: String(data.get("email") ?? "").trim(),
    };

    const result = await apiFetch("/api/auth/password/reset", {
      method: "POST",
      body: JSON.stringify(payload),
    });

    setLoading(false);

    if (result.status === 429) {
      toast.error(t("rateLimitedToast"));
      return;
    }

    if (!result.ok) {
      toast.error(t("failureToast"));
      return;
    }

    toast.success(t("successToast"));
    form.reset();
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
                <FieldLabel htmlFor="email">{t("emailLabel")}</FieldLabel>
                <Input
                  id="email"
                  type="email"
                  name="email"
                  placeholder="m@example.com"
                  required
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
