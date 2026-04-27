import Link from "next/link";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import QuoteListTable from "@/components/quote/quote-list-table";

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
          <Button asChild>
            <Link href="/quote/create">Nouveau devis</Link>
          </Button>
        </CardHeader>
        <CardContent>
          <QuoteListTable />
        </CardContent>
      </Card>
    </>
  );
}
