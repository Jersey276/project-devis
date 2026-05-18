export const SUPPORTED_LOCALES = ["fr"] as const;
export const DEFAULT_LOCALE: AppLocale = "fr";
export type AppLocale = (typeof SUPPORTED_LOCALES)[number];

export function isSupportedLocale(value: string): value is AppLocale {
  return (SUPPORTED_LOCALES as readonly string[]).includes(value);
}
