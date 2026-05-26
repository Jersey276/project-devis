import { getTranslations } from "next-intl/server";

export default async function InvoiceIndexPage() {
  const t = await getTranslations("invoice.list");
  return <div>{t("placeholder")}</div>;
}
