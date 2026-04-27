import Link from "next/link";
import PageBreadcrumb from "@/components/custom/page-breadcrumb";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  DataTable,
  DataTableBody,
  DataTableCell,
  DataTableHead,
  DataTableHeader,
  DataTableRow,
  DataTableRowAction,
  DataTableRowActions,
  DataTableSortableHead,
} from "@/components/custom/data-table";
import { computeQuoteTotal, quoteMocks } from "@/lib/mocks/quotes";

type QuoteListItem = {
  id: string;
  projectName: string;
  status: "draft" | "sent" | "signed";
  total: number;
};

const data: QuoteListItem[] = quoteMocks.map((quote) => ({
  id: quote.uuid,
  projectName: quote.name,
  status: quote.status,
  total: computeQuoteTotal(quote),
}));

const rowActions: DataTableRowAction[] = [
  {
    type: "link",
    label: "Voir/Modifier",
    href: "/quote/{id}",
  },
];

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
          <DataTable
            datas={data}
            sortBy="id"
            sortDirection="asc"
            row_actions={rowActions}
          >
            <DataTableHeader>
              <DataTableRow>
                <DataTableSortableHead name="id">ID</DataTableSortableHead>
                <DataTableSortableHead name="projectName">
                  Projet
                </DataTableSortableHead>
                <DataTableSortableHead name="status">
                  Statut
                </DataTableSortableHead>
                <DataTableSortableHead name="total">
                  Total
                </DataTableSortableHead>
                <DataTableHead>Actions</DataTableHead>
              </DataTableRow>
            </DataTableHeader>
            <DataTableBody>
              {data.map((quote) => (
                <DataTableRow key={quote.id}>
                  <DataTableCell>{quote.id}</DataTableCell>
                  <DataTableCell>{quote.projectName}</DataTableCell>
                  <DataTableCell>{quote.status}</DataTableCell>
                  <DataTableCell>{quote.total.toFixed(2)} €</DataTableCell>
                  <DataTableCell>
                    <DataTableRowActions id={quote.id} />
                  </DataTableCell>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
        </CardContent>
      </Card>
    </>
  );
}
