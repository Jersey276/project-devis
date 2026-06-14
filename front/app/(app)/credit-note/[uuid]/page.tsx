import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import SubscriptionGuard from "@/components/custom/subscription-guard";
import CreditNoteDetail from "@/components/invoice/credit-note-detail";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function CreditNoteDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const tInvoice = await getTranslations("invoice.list");
  const t = await getTranslations("creditNote.detail");
  const breadcrumbs = [
    { href: "/invoice", label: tInvoice("title") },
    { href: `/credit-note/${uuid}`, label: t("breadcrumb") },
  ];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <SubscriptionGuard>
        <CreditNoteDetail creditNoteId={uuid} />
      </SubscriptionGuard>
    </>
  );
}
