import { getTranslations } from "next-intl/server";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import QuoteListTable from "@/components/quote/quote-list-table";
import NewQuoteButton from "@/components/quote/new-quote-button";

export default async function QuoteIndexPage() {
  const t = await getTranslations("quote");
  const breadcrumbs = [
    {
      href: "/quote",
      label: t("breadcrumb.list"),
    },
  ];
  return (
    <>
      <PageBreadcrumb items={breadcrumbs} />
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle>{t("list.title")}</CardTitle>
          <NewQuoteButton />
        </CardHeader>
        <CardContent>
          <QuoteListTable />
        </CardContent>
      </Card>
    </>
  );
}
