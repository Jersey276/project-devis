import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import QuoteForm from "@/components/quote/quote-form";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function QuoteDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const t = await getTranslations("quote.breadcrumb");

  const breadcrumbs = [
    {
      href: "/quote",
      label: t("list"),
    },
    {
      href: `/quote/${uuid}`,
      label: t("detail"),
    },
  ];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <QuoteForm quoteId={uuid} />
    </>
  );
}
