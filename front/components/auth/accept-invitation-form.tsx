"use client";

import { FormEvent, useEffect, useMemo, useState } from "react";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const CODE_INVALID_TOKEN = 1016;
const CODE_EXPIRED_TOKEN = 1017;
const CODE_ALREADY_LINKED = 1018;

export default function AcceptInvitationForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const t = useTranslations("auth.invite");
  const params = useSearchParams();
  const router = useRouter();
  const token = useMemo(() => params.get("token")?.trim() ?? "", [params]);

  const [isLoggedIn, setIsLoggedIn] = useState<boolean | null>(null);
  const [loading, setLoading] = useState(false);
  const [confirmError, setConfirmError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch("/api/auth/me").then(({ ok, body }) => {
      setIsLoggedIn(ok && body.success === true);
    });
  }, []);

  function setCustomerModeCookie() {
    document.cookie = "user-mode=customer; path=/; max-age=31536000; SameSite=Lax";
  }

  function handleTokenError(code: number) {
    if (code === CODE_INVALID_TOKEN || code === CODE_ALREADY_LINKED) {
      toast.error(t("alreadyLinked"));
    } else if (code === CODE_EXPIRED_TOKEN) {
      toast.error(t("expiredLink"));
    } else {
      toast.error(t("invalidLink"));
    }
  }

  function onSuccess() {
    setCustomerModeCookie();
    toast.success(t("acceptSuccess"));
    router.replace("/client-profile");
  }

  async function handleLinkExisting() {
    setLoading(true);
    try {
      const { ok, body } = await apiFetch("/api/auth/invite/accept-linked", {
        method: "POST",
        body: JSON.stringify({ token }),
      });
      if (ok && body.success) {
        onSuccess();
      } else {
        handleTokenError(body.code as number);
      }
    } catch {
      toast.error(t("invalidLink"));
    } finally {
      setLoading(false);
    }
  }

  async function handleLoginAndLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    const form = e.currentTarget;
    const data = new FormData(form);
    const email = String(data.get("email") ?? "");
    const password = String(data.get("password") ?? "");

    setLoading(true);
    try {
      const loginRes = await apiFetch("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      if (!loginRes.ok || !loginRes.body.success) {
        toast.error(t("invalidLink"));
        return;
      }

      const linkRes = await apiFetch("/api/auth/invite/accept-linked", {
        method: "POST",
        body: JSON.stringify({ token }),
      });
      if (linkRes.ok && linkRes.body.success) {
        onSuccess();
      } else {
        handleTokenError(linkRes.body.code as number);
      }
    } catch {
      toast.error(t("invalidLink"));
    } finally {
      setLoading(false);
    }
  }

  async function handleRegisterAndLink(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setConfirmError(null);
    const form = e.currentTarget;
    const data = new FormData(form);
    const email = String(data.get("email") ?? "");
    const password = String(data.get("password") ?? "");
    const confirmPassword = String(data.get("confirm_password") ?? "");

    if (password !== confirmPassword) {
      setConfirmError(t("invalidLink"));
      return;
    }

    setLoading(true);
    try {
      const { ok, body } = await apiFetch("/api/auth/invite/accept", {
        method: "POST",
        body: JSON.stringify({ token, email, password }),
      });
      if (ok && body.success) {
        onSuccess();
      } else {
        handleTokenError(body.code as number);
      }
    } catch {
      toast.error(t("invalidLink"));
    } finally {
      setLoading(false);
    }
  }

  if (!token) {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader>
            <CardTitle>{t("invalidLinkTitle")}</CardTitle>
            <CardDescription>{t("invalidLinkDescription")}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (isLoggedIn === null) {
    return null;
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>{t("acceptTitle")}</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoggedIn ? (
            <div className="flex flex-col gap-4">
              <p className="text-sm text-muted-foreground">
                {t("linkButton")}
              </p>
              <Button onClick={handleLinkExisting} disabled={loading}>
                {loading ? "…" : t("linkButton")}
              </Button>
            </div>
          ) : (
            <Tabs defaultValue="login">
              <TabsList className="mb-4 w-full">
                <TabsTrigger value="login" className="flex-1">
                  {t("acceptLoginTab")}
                </TabsTrigger>
                <TabsTrigger value="register" className="flex-1">
                  {t("acceptRegisterTab")}
                </TabsTrigger>
              </TabsList>

              <TabsContent value="login">
                <form onSubmit={handleLoginAndLink} noValidate>
                  <FieldGroup>
                    <Field>
                      <FieldLabel htmlFor="login-email">{t("loginLabel")}</FieldLabel>
                      <Input
                        id="login-email"
                        type="email"
                        name="email"
                        autoComplete="email"
                        required
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="login-password">{t("passwordLabel")}</FieldLabel>
                      <Input
                        id="login-password"
                        type="password"
                        name="password"
                        autoComplete="current-password"
                        required
                      />
                    </Field>
                    <Field>
                      <Button type="submit" disabled={loading}>
                        {loading ? "…" : t("acceptLoginSubmit")}
                      </Button>
                    </Field>
                  </FieldGroup>
                </form>
              </TabsContent>

              <TabsContent value="register">
                <form onSubmit={handleRegisterAndLink} noValidate>
                  <FieldGroup>
                    <Field>
                      <FieldLabel htmlFor="reg-email">{t("emailLabel")}</FieldLabel>
                      <Input
                        id="reg-email"
                        type="email"
                        name="email"
                        autoComplete="email"
                        required
                      />
                    </Field>
                    <Field>
                      <FieldLabel htmlFor="reg-password">{t("newPasswordLabel")}</FieldLabel>
                      <Input
                        id="reg-password"
                        type="password"
                        name="password"
                        autoComplete="new-password"
                        required
                      />
                      <FieldDescription>{t("newPasswordHint")}</FieldDescription>
                    </Field>
                    <Field data-invalid={!!confirmError}>
                      <FieldLabel htmlFor="reg-confirm">{t("newPasswordLabel")}</FieldLabel>
                      <Input
                        id="reg-confirm"
                        type="password"
                        name="confirm_password"
                        autoComplete="new-password"
                        required
                        aria-invalid={!!confirmError}
                      />
                      <FieldError
                        errors={confirmError ? [{ message: confirmError }] : undefined}
                      />
                    </Field>
                    <Field>
                      <Button type="submit" disabled={loading}>
                        {loading ? "…" : t("acceptRegisterSubmit")}
                      </Button>
                    </Field>
                  </FieldGroup>
                </form>
              </TabsContent>
            </Tabs>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
