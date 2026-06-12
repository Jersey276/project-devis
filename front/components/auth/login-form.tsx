"use client";
import { FormEvent } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { NEXT_PARAM, safeNextPath } from "@/lib/auth-utils";
import { useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";

function submitLoginForm(
  router: ReturnType<typeof useRouter>,
  next: string,
  messages: { success: string; failure: string },
) {
  return async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);
    const email = data.get("email");
    const password = data.get("password");
    const rememberMe = data.get("remember_me") === "on";
    await fetch("/api/auth/login", {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password, remember_me: rememberMe }),
    })
      .then(async (response) => {
        if (response.ok) {
          toast.success(messages.success);
          router.replace(next);
        } else {
          toast.error(messages.failure);
        }
      })
      .catch(() => {
        toast.error(messages.failure);
      });
  };
}

export default function LoginForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const router = useRouter();
  const next = safeNextPath(useSearchParams().get(NEXT_PARAM));
  const t = useTranslations("auth.login");
  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <form
            onSubmit={submitLoginForm(router, next, {
              success: t("successToast"),
              failure: t("failureToast"),
            })}
          >
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
                <div className="flex items-center">
                  <FieldLabel htmlFor="password">
                    {t("passwordLabel")}
                  </FieldLabel>
                  <a
                    href="/forget-password"
                    className="ml-auto inline-block text-sm underline-offset-4 hover:underline"
                  >
                    {t("forgotLink")}
                  </a>
                </div>
                <Input id="password" type="password" name="password" required />
              </Field>
              <Field>
                <div className="flex items-center gap-2">
                  <Checkbox id="remember_me" name="remember_me" />
                  <FieldLabel htmlFor="remember_me" className="font-normal">
                    {t("rememberMeLabel")}
                  </FieldLabel>
                </div>
              </Field>
              <Field>
                <Button type="submit">{t("submit")}</Button>
                <Button variant="outline" type="button">
                  {t("googleSubmit")}
                </Button>
                <FieldDescription className="text-center">
                  {t("signupPrompt")} <a href="/register">{t("signupLink")}</a>
                </FieldDescription>
              </Field>
            </FieldGroup>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
