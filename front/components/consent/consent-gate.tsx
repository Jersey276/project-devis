"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { toast } from "sonner";
import { CONSENT_VERSIONS, type ConsentType } from "@/lib/consent-versions";
import { Checkbox } from "@/components/ui/checkbox";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";

type Props = {
  outdated: ConsentType[];
};

export default function ConsentGate({ outdated }: Props) {
  const t = useTranslations("auth.consentGate");
  const [accepted, setAccepted] = useState<Record<ConsentType, boolean>>({
    cgv: false,
    privacy_policy: false,
  });
  const [loading, setLoading] = useState(false);
  const [done, setDone] = useState(false);

  if (done || outdated.length === 0) return null;

  const allChecked = outdated.every((type) => accepted[type]);

  async function handleSubmit() {
    if (!allChecked) return;
    setLoading(true);
    try {
      for (const type of outdated) {
        const res = await fetch("/api/consent", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({ type, version: CONSENT_VERSIONS[type] }),
        });
        if (!res.ok) throw new Error();
      }
      setDone(true);
    } catch {
      toast.error(t("errorToast"));
    } finally {
      setLoading(false);
    }
  }

  return (
    <Dialog open={!done}>
      <DialogContent
        className="sm:max-w-md"
        // Empêche la fermeture par clic extérieur ou touche Escape
        onInteractOutside={(e) => e.preventDefault()}
        onEscapeKeyDown={(e) => e.preventDefault()}
      >
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
          <DialogDescription>{t("description")}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-2">
          {outdated.includes("cgv") && (
            <div className="flex items-start gap-2">
              <Checkbox
                id="gate-cgv"
                checked={accepted.cgv}
                onCheckedChange={(v) =>
                  setAccepted((prev) => ({ ...prev, cgv: v === true }))
                }
              />
              <label htmlFor="gate-cgv" className="text-sm leading-snug cursor-pointer">
                {t.rich("cgvLabel", {
                  cgvLink: (chunks) => <Link href="/cgv" target="_blank" className="underline">{chunks}</Link>,
                })}
              </label>
            </div>
          )}
          {outdated.includes("privacy_policy") && (
            <div className="flex items-start gap-2">
              <Checkbox
                id="gate-privacy"
                checked={accepted.privacy_policy}
                onCheckedChange={(v) =>
                  setAccepted((prev) => ({ ...prev, privacy_policy: v === true }))
                }
              />
              <label htmlFor="gate-privacy" className="text-sm leading-snug cursor-pointer">
                {t.rich("privacyLabel", {
                  privacyLink: (chunks) => <Link href="/politique-de-confidentialite" target="_blank" className="underline">{chunks}</Link>,
                })}
              </label>
            </div>
          )}
        </div>

        <Button onClick={handleSubmit} disabled={!allChecked || loading}>
          {t("submit")}
        </Button>
      </DialogContent>
    </Dialog>
  );
}
