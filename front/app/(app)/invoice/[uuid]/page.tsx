import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import InvoiceDetail from "@/components/invoice/invoice-detail";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function InvoiceDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const t = await getTranslations("invoice.list");
  const breadcrumbs = [
    { href: "/invoice", label: t("title") },
    { href: `/invoice/${uuid}`, label: uuid },
  ];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <SubscriptionGuard>
        <InvoiceDetail invoiceId={uuid} />
      </SubscriptionGuard>
    </>
  );
}
