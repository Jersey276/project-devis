import { getTranslations } from "next-intl/server";
import Link from "next/link";

export default async function MentionsLegalesPage() {
  const t = await getTranslations("legal");
  const tm = await getTranslations("legal.mentions");

  return (
    <article className="prose prose-neutral max-w-none dark:prose-invert">
      <Link href="/" className="not-prose text-sm text-muted-foreground hover:underline">
        ← {t("backHome")}
      </Link>

      <h1 className="mt-6">{tm("title")}</h1>
      <p className="text-sm text-muted-foreground">
        {t("lastUpdated", { date: "2025-07-01" })}
      </p>

      <h2>{tm("s1Title")}</h2>
      <p style={{ whiteSpace: "pre-line" }}>{tm("s1")}</p>

      <h2>{tm("s2Title")}</h2>
      <p>{tm("s2")}</p>

      <h2>{tm("s3Title")}</h2>
      <p>{tm("s3")}</p>

      <h2>{tm("s4Title")}</h2>
      <p>{tm("s4")}</p>

      <h2>{tm("s5Title")}</h2>
      <p>{tm("s5")}</p>
    </article>
  );
}
