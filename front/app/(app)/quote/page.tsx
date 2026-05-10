import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import QuoteListTable from "@/components/quote/quote-list-table";
import NewQuoteButton from "@/components/quote/new-quote-button";

const breadcrumbs = [
  {
    href: "/quote",
    label: "Devis",
  },
];

export default function QuoteIndexPage() {
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>Devis</CardTitle>
          <NewQuoteButton />
        </CardHeader>
        <CardContent>
          <QuoteListTable />
        </CardContent>
      </Card>
    </>
  );
}
