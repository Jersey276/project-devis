import { getTranslations } from "next-intl/server";
import Link from "next/link";

export default async function CgvPage() {
  const t = await getTranslations("legal");
  const tc = await getTranslations("legal.cgv");

  return (
    <article className="prose prose-neutral max-w-none dark:prose-invert">
      <Link href="/" className="not-prose text-sm text-muted-foreground hover:underline">
        ← {t("backHome")}
      </Link>

      <h1 className="mt-6">{tc("title")}</h1>
      <p className="text-sm text-muted-foreground">
        {t("lastUpdated", { date: "2025-07-01" })}
      </p>

      <h2>{tc("s1Title")}</h2>
      <p>{tc("s1")}</p>

      <h2>{tc("s2Title")}</h2>
      <p>{tc("s2")}</p>

      <h2>{tc("s3Title")}</h2>
      <p>{tc("s3")}</p>

      <h2>{tc("s4Title")}</h2>
      <p>{tc("s4")}</p>

      <h2>{tc("s5Title")}</h2>
      <p>{tc("s5")}</p>

      <h2>{tc("s6Title")}</h2>
      <p>{tc("s6")}</p>

      <h2>{tc("s7Title")}</h2>
      <p>{tc("s7")}</p>
    </article>
  );
}
