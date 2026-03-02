import { AppLayout } from "@/app/layout";
import QuoteForm from "@/components/quote/quote-form";
import { getQuoteMockById } from "@/lib/mocks/quotes";
import { redirect } from "next/navigation";

type PageProps = {
  params: Promise<{
    uuid: string;
  }>;
};

export default async function QuoteDetailPage({ params }: PageProps) {
  const { uuid } = await params;
  const existingQuote = getQuoteMockById(uuid);

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

  if (!existingQuote) {
    redirect("/quote");
  }
  const quote = existingQuote;

  return (
    <AppLayout breadcrumbs={breadcrumbs}>
      <QuoteForm quote={quote} />
    </AppLayout>
  );
}
