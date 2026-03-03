import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { AppLayout } from "../layout";
import { Button } from "@/components/ui/button";
import Link from "next/link";
import { PlusIcon } from "lucide-react";
import { ClientsTable } from "./clients-table";

interface Clients {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
}

const data: Clients[] = [
  {
    id: "m5gr84i9",
    first_name: "John",
    last_name: "Doe",
    email: "john.doe@example.com",
  },
  {
    id: "3u1reuv4",
    first_name: "Jane",
    last_name: "Smith",
    email: "jane.smith@example.com",
  },
  {
    id: "derv1ws0",
    first_name: "Monserrat",
    last_name: "Lopez",
    email: "monserrat@example.com",
  },
  {
    id: "5kma53ae",
    first_name: "Silas",
    last_name: "Brown",
    email: "silas.brown@example.com",
  },
  {
    id: "bhqecj4p",
    first_name: "Carmella",
    last_name: "Rossi",
    email: "carmella@example.com",
  },
];

export default function ClientIndex() {
  return (
    <AppLayout>
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between gap-4">
            <CardTitle>Clients</CardTitle>
            <Button asChild>
              <Link
                href="/clients/create"
                className="inline-flex items-center gap-2"
              >
                <PlusIcon className="h-4 w-4" />
                Nouveau client
              </Link>
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <ClientsTable data={data} />
        </CardContent>
      </Card>
    </AppLayout>
  );
}
