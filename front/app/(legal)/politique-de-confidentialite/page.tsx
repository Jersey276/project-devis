import { getTranslations } from "next-intl/server";
import Link from "next/link";

export default async function PolitiqueConfidentialitePage() {
  const t = await getTranslations("legal");
  const tp = await getTranslations("legal.privacy");

  return (
    <article className="prose prose-neutral max-w-none dark:prose-invert">
      <Link href="/" className="not-prose text-sm text-muted-foreground hover:underline">
        ← {t("backHome")}
      </Link>

      <h1 className="mt-6">{tp("title")}</h1>
      <p className="text-sm text-muted-foreground">
        {t("lastUpdated", { date: "2025-07-01" })}
      </p>

      <h2>{tp("s1Title")}</h2>
      <p>{tp("s1")}</p>

      <h2>{tp("s2Title")}</h2>
      <p style={{ whiteSpace: "pre-line" }}>{tp("s2")}</p>

      <h2>{tp("s3Title")}</h2>
      <p>{tp("s3")}</p>

      <h2>{tp("s4Title")}</h2>
      <p>{tp("s4")}</p>

      <h2>{tp("s5Title")}</h2>
      <p>{tp("s5")}</p>

      <h2>{tp("s6Title")}</h2>
      <p>{tp("s6")}</p>

      <h2>{tp("s7Title")}</h2>
      <p>{tp("s7")}</p>
    </article>
  );
}
