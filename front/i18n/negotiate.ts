import { DEFAULT_LOCALE, SUPPORTED_LOCALES, type AppLocale } from "./locales";

type Candidate = { tag: string; quality: number };

function parseAcceptLanguage(header: string): Candidate[] {
  return header
    .split(",")
    .map((part) => {
      const [rawTag, ...params] = part.trim().split(";");
      const tag = rawTag.trim().toLowerCase();
      if (!tag) return null;
      const qParam = params.find((p) => p.trim().startsWith("q="));
      const quality = qParam ? Number.parseFloat(qParam.split("=")[1]) : 1;
      if (!Number.isFinite(quality) || quality <= 0) return null;
      return { tag, quality };
    })
    .filter((c): c is Candidate => c !== null)
    .sort((a, b) => b.quality - a.quality);
}

// Match either an exact locale ("fr") or the primary subtag ("fr-FR" → "fr").
function matchLocale(tag: string): AppLocale | null {
  const lower = tag.toLowerCase();
  for (const locale of SUPPORTED_LOCALES) {
    if (lower === locale || lower.startsWith(`${locale}-`)) return locale;
  }
  return null;
}

export function negotiateLocale(header: string | null | undefined): AppLocale {
  // No need to parse the header when only one locale is supported.
  if (SUPPORTED_LOCALES.length === 1) return DEFAULT_LOCALE;
  if (!header) return DEFAULT_LOCALE;
  for (const candidate of parseAcceptLanguage(header)) {
    const matched = matchLocale(candidate.tag);
    if (matched) return matched;
  }
  return DEFAULT_LOCALE;
}
