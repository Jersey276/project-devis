import { headers } from "next/headers";
import { getRequestConfig } from "next-intl/server";
import { negotiateLocale } from "./negotiate";

export default getRequestConfig(async () => {
  // Cookie override (e.g. user-selected locale) will plug in here when a
  // language switcher ships. Today: browser Accept-Language only.
  const headerStore = await headers();
  const locale = negotiateLocale(headerStore.get("accept-language"));
  const messages = (await import(`../messages/${locale}.json`)).default;
  return {
    locale,
    messages,
    timeZone: "Europe/Paris",
  };
});
