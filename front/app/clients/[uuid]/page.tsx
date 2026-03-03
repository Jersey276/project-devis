"use client";

import { AppLayout } from "@/app/layout";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useParams } from "next/navigation";

type ClientAddress = {
  id: string;
  label: string;
  line: string;
  city: string;
  zipCode: string;
  country: string;
};

const addresses: ClientAddress[] = [
  {
    id: "addr-1",
    label: "Adresse principale",
    line: "12 Rue des Lilas",
    city: "Paris",
    zipCode: "75001",
    country: "France",
  },
  {
    id: "addr-2",
    label: "Adresse facturation",
    line: "2 Avenue de la République",
    city: "Lyon",
    zipCode: "69002",
    country: "France",
  },
  {
    id: "addr-3",
    label: "Adresse livraison",
    line: "8 Rue du Port",
    city: "Bordeaux",
    zipCode: "33000",
    country: "France",
  },
];

export default function ClientProfilePage() {
  const { uuid } = useParams<{ uuid: string }>();
  const client = {
    id: uuid,
    firstName: "John",
    lastName: "Doe",
    email: "john.doe@example.com",
  };

  return (
    <AppLayout>
      <Card className="gap-4">
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="space-y-1">
              <CardTitle>Profil client</CardTitle>
              <CardDescription>
                ID: {client.id} • {client.firstName} {client.lastName}
              </CardDescription>
            </div>
            <Badge variant="outline">Client</Badge>
          </div>
        </CardHeader>

        <CardContent className="grid gap-4">
          <Card size="sm" className="gap-3">
            <CardHeader className="pb-0">
              <CardTitle>Informations</CardTitle>
            </CardHeader>
            <CardContent className="space-y-1 text-sm">
              <p>
                <span className="font-medium">Prénom:</span> {client.firstName}
              </p>
              <p>
                <span className="font-medium">Nom:</span> {client.lastName}
              </p>
              <p>
                <span className="font-medium">Email:</span> {client.email}
              </p>
            </CardContent>
          </Card>

          <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
            {addresses.map((address) => (
              <Card key={address.id} size="sm" className="gap-3">
                <CardHeader className="pb-0">
                  <CardTitle>{address.label}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-1 text-sm">
                  <p>{address.line}</p>
                  <p>
                    {address.zipCode} {address.city}
                  </p>
                  <p>{address.country}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>
    </AppLayout>
  );
}
