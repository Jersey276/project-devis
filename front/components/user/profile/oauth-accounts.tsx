"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import OAuthButtons, {
  type OAuthProvider,
} from "@/components/auth/oauth-buttons";
import { apiFetch } from "@/lib/api";

const PROVIDER_LABELS: Record<OAuthProvider, string> = {
  google: "Google",
  github: "GitHub",
  microsoft: "Microsoft",
};

type LinkedIdentity = {
  provider: OAuthProvider;
  email: string;
};

type OAuthAccountsProps = {
  readOnly?: boolean;
};

/**
 * OAuthAccounts lists the OAuth providers linked to the current account and lets
 * the user link a new one or unlink an existing one. The last remaining login
 * method (no password + a single identity) cannot be unlinked — guarded both
 * here and by the backend.
 */
export default function OAuthAccounts({ readOnly = false }: OAuthAccountsProps) {
  const t = useTranslations("profile.connection");
  const tCommon = useTranslations("common");
  const tOAuthErrors = useTranslations("auth.oauth.errors");
  const searchParams = useSearchParams();
  const [identities, setIdentities] = useState<LinkedIdentity[]>([]);
  const [hasPassword, setHasPassword] = useState(true);
  const [loading, setLoading] = useState(true);
  const [pending, setPending] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/auth/oauth-identities").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setIdentities((body.identities as LinkedIdentity[]) ?? []);
        setHasPassword(Boolean(body.has_password));
      } else {
        toast.error(t("oauthLoadError"));
      }
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
    // Fetch once on mount; `t` is read at call time and intentionally excluded
    // to avoid refetching when the translations function identity changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Surface the OAuth link outcome redirected back here by the gateway callback
  // as ?oauth_linked=<provider> (success) or ?oauth_error=<slug> (failure).
  const oauthLinked = searchParams.get("oauth_linked");
  const oauthError = searchParams.get("oauth_error");
  useEffect(() => {
    if (oauthError) {
      toast.error(tOAuthErrors(oauthError as never));
    } else if (oauthLinked) {
      toast.success(t("oauthLinkedToast"));
    }
    // `t`/`tOAuthErrors` are read at call time and intentionally excluded.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [oauthLinked, oauthError]);

  const isLastMethod = !hasPassword && identities.length <= 1;

  async function handleUnlink(provider: OAuthProvider) {
    if (readOnly || isLastMethod) return;
    setPending(provider);
    try {
      const { ok, body } = await apiFetch(
        `/api/auth/oauth-identities/${provider}`,
        { method: "DELETE" },
      );
      if (ok && body.success) {
        setIdentities((prev) => prev.filter((i) => i.provider !== provider));
        toast.success(t("oauthUnlinkedToast"));
        return;
      }
      toast.error(body.message ?? tCommon("errors.generic"));
    } catch {
      toast.error(tCommon("errors.generic"));
    } finally {
      setPending(null);
    }
  }

  if (loading) {
    return (
      <p className="text-muted-foreground text-sm">
        {tCommon("actions.loading")}
      </p>
    );
  }

  return (
    <div className="grid max-w-3xl gap-4">
      <div className="space-y-1">
        <h3 className="text-sm font-medium">{t("oauthTitle")}</h3>
        <p className="text-muted-foreground text-sm">{t("oauthHint")}</p>
      </div>

      {identities.length > 0 ? (
        <ul className="divide-y rounded-md border">
          {identities.map((identity) => (
            <li
              key={identity.provider}
              className="flex items-center justify-between gap-3 px-3 py-2"
            >
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">
                  {PROVIDER_LABELS[identity.provider] ?? identity.provider}
                </span>
                <Badge variant="outline">{t("linkedLabel")}</Badge>
                <span className="text-muted-foreground text-sm">
                  {identity.email}
                </span>
              </div>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                disabled={readOnly || isLastMethod || pending === identity.provider}
                onClick={() => handleUnlink(identity.provider)}
              >
                {pending === identity.provider
                  ? tCommon("actions.saving")
                  : t("disconnect")}
              </Button>
            </li>
          ))}
        </ul>
      ) : null}

      {isLastMethod ? (
        <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
          {t("lastMethodWarning")}
        </p>
      ) : null}

      {!readOnly ? (
        <div className="grid gap-2">
          <OAuthButtons
            mode="link"
            next="/profile?tab=compte"
            linked={identities.map((i) => i.provider)}
          />
        </div>
      ) : null}
    </div>
  );
}
