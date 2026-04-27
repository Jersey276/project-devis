import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import QuoteForm from "@/components/quote/quote-form";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function QuoteDetailPage({ params }: PageProps) {
  const { uuid } = await params;

  const breadcrumbs = [
    {
      href: "/quote",
      label: "Devis",
    },
    {
      href: `/quote/${uuid}`,
      label: "Détail du devis",
    },
  ];

  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <QuoteForm quoteId={uuid} />
    </>
  );
}
