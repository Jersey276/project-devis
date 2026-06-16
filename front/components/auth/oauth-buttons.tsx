"use client";

import { useTranslations } from "next-intl";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import OAuthIcon from "@/components/auth/oauth-icons";

const PROVIDERS = ["google", "github", "microsoft"] as const;
export type OAuthProvider = (typeof PROVIDERS)[number];

type OAuthButtonsProps = {
  /** Post-flow landing path, carried through the OAuth round-trip. */
  next: string;
  /**
   * "login" (default) starts the public sign-in/sign-up flow.
   * "link" attaches a provider to the already-authenticated account.
   */
  mode?: "login" | "link";
  /** In link mode, providers already linked are hidden. */
  linked?: readonly string[];
};

/**
 * OAuthButtons renders one full-page navigation link per configured provider.
 * OAuth requires a top-level redirect (not fetch), so these are anchors styled
 * as buttons. `next` is the post-flow landing path, carried through the flow.
 */
export default function OAuthButtons({
  next,
  mode = "login",
  linked = [],
}: OAuthButtonsProps) {
  const t = useTranslations("auth.oauth");
  const basePath = mode === "link" ? "/api/auth/oauth-link" : "/api/auth/oauth";
  const providers =
    mode === "link" ? PROVIDERS.filter((p) => !linked.includes(p)) : PROVIDERS;

  return (
    <>
      {providers.map((provider) => (
        <a
          key={provider}
          href={`${basePath}/${provider}?next=${encodeURIComponent(next)}`}
          className={cn(buttonVariants({ variant: "outline" }), "w-full gap-2")}
        >
          <OAuthIcon provider={provider} className="size-4" />
          {mode === "link" ? t(`connect.${provider}`) : t(provider)}
        </a>
      ))}
    </>
  );
}
