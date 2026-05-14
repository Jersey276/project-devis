import { getTranslations } from "next-intl/server";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function InvoiceDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const t = await getTranslations("invoice.detail");
  return <div>{t("title", { id: uuid })}</div>;
}
