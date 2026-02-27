import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AppLayout } from "../layout";
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
import { PencilIcon } from "lucide-react";

interface Clients {
  id: string;
  name: string;
  email: string;
  nb_active_projects?: number;
}

const data: Clients[] = [
  {
    id: "m5gr84i9",
    name: "John Doe",
    email: "ken99@example.com",
    nb_active_projects: 3,
  },
  {
    id: "3u1reuv4",
    name: "Jane Smith",
    email: "Abe45@example.com",
    nb_active_projects: 5,
  },
  {
    id: "derv1ws0",
    name: "Monserrat",
    email: "Monserrat44@example.com",
    nb_active_projects: 2,
  },
  {
    id: "5kma53ae",
    name: "Silas",
    email: "Silas22@example.com",
    nb_active_projects: 1,
  },
  {
    id: "bhqecj4p",
    name: "Carmella",
    email: "carmella@example.com",
    nb_active_projects: 0,
  },
];

const EditActionIcon = () => <PencilIcon className="h-4 w-4" />;

const row_actions: DataTableRowAction[] = [
  {
    type: "link",
    label: "Voir/Modifier",
    icon: EditActionIcon,
    href: "/edit/{id}",
  },
];

export default function clientIndex() {
  const sortBy = "id";
  const sortDirection = "asc";

  return (
    <AppLayout>
      <Card>
        <CardHeader>
          <CardTitle>Clients</CardTitle>
        </CardHeader>
        <CardContent>
          <DataTable
            datas={data}
            sortBy={sortBy}
            sortDirection={sortDirection}
            row_actions={row_actions}
          >
            <DataTableHeader>
              <DataTableRow>
                <DataTableSortableHead name="id">ID</DataTableSortableHead>
                <DataTableSortableHead name="name">Name</DataTableSortableHead>
                <DataTableSortableHead name="email">
                  Email
                </DataTableSortableHead>
                <DataTableSortableHead name="nb_active_projects">
                  Projet Actifs
                </DataTableSortableHead>
                <DataTableHead>Actions</DataTableHead>
              </DataTableRow>
            </DataTableHeader>
            <DataTableBody>
              {data.map((client, index) => (
                <DataTableRow key={index}>
                  <DataTableCell>{client.id}</DataTableCell>
                  <DataTableCell>{client.name}</DataTableCell>
                  <DataTableCell>{client.email}</DataTableCell>
                  <DataTableCell className="text-right">
                    {client.nb_active_projects}
                  </DataTableCell>
                  <DataTableCell>
                    <DataTableRowActions id={client.id} />
                  </DataTableCell>
                </DataTableRow>
              ))}
            </DataTableBody>
          </DataTable>
        </CardContent>
      </Card>
    </AppLayout>
  );
}
