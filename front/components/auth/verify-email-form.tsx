"use client";

import { useEffect, useMemo, useState } from "react";
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
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";
import { toast } from "sonner";

const CODE_INVALID_VERIFICATION_TOKEN = 1010;
const CODE_EXPIRED_VERIFICATION_TOKEN = 1011;
const CODE_ALREADY_VERIFIED = 1012;

export default function VerifyEmailForm({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  const t = useTranslations("auth.verifyEmail");
  const params = useSearchParams();
  const router = useRouter();
  const token = useMemo(() => params.get("token")?.trim() ?? "", [params]);

  const [verifying, setVerifying] = useState(!!token);
  const [verifyDone, setVerifyDone] = useState(false);
  const [resendLoading, setResendLoading] = useState(false);

  useEffect(() => {
    if (!token) return;

    let cancelled = false;
    (async () => {
      const result = await apiFetch("/api/auth/email/verify", {
        method: "POST",
        body: JSON.stringify({ token }),
      });
      if (cancelled) return;
      setVerifying(false);
      if (result.ok || result.body.code === CODE_ALREADY_VERIFIED) {
        setVerifyDone(true);
        toast.success(t("verifySuccess"));
        return;
      }
      if (result.body.code === CODE_INVALID_VERIFICATION_TOKEN) {
        toast.error(t("verifyInvalidToken"));
        return;
      }
      if (result.body.code === CODE_EXPIRED_VERIFICATION_TOKEN) {
        toast.error(t("verifyExpiredToken"));
        return;
      }
      toast.error(t("verifyInvalidToken"));
    })();
    return () => {
      cancelled = true;
    };
  }, [token, t]);

  async function handleResend() {
    setResendLoading(true);
    const result = await apiFetch("/api/auth/email/resend-verification", {
      method: "POST",
    });
    setResendLoading(false);

    if (result.status === 429) {
      toast.error(t("resendRateLimited"));
      return;
    }
    if (result.ok) {
      toast.success(t("resendSuccess"));
      return;
    }
    toast.error(t("resendError"));
  }

  if (token && verifying) {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader>
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>{t("resendLoading")}</CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  if (verifyDone) {
    return (
      <div className={cn("flex flex-col gap-6", className)} {...props}>
        <Card>
          <CardHeader>
            <CardTitle>{t("verifySuccess")}</CardTitle>
          </CardHeader>
          <CardContent>
            <Button type="button" onClick={() => router.replace("/")}>
              {t("backToApp")}
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className={cn("flex flex-col gap-6", className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
          <CardDescription>{t("description")}</CardDescription>
        </CardHeader>
        <CardContent>
          <Button
            type="button"
            onClick={handleResend}
            disabled={resendLoading}
          >
            {resendLoading ? t("resendLoading") : t("resendButton")}
          </Button>
        </CardContent>
      </Card>
    </div>
  );
}
