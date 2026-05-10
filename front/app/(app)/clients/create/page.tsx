"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import ClientForm, {
  EMPTY_CLIENT_VALUES,
  type ClientFormValues,
} from "@/components/user/client/client-form";
import {
  EMPTY_ADDRESS_VALUES,
  type AddressValues,
} from "@/components/address/address-form";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { createClient } from "@/lib/services/clients";
import { createAddress } from "@/lib/services/addresses";

export default function CreateClientPage() {
  const router = useRouter();
  const [client, setClient] = useState<ClientFormValues>(EMPTY_CLIENT_VALUES);
  const [address, setAddress] = useState<AddressValues>(EMPTY_ADDRESS_VALUES);
  const [submitting, setSubmitting] = useState(false);
  const [clientErrors, setClientErrors] = useState<FieldErrors>({});
  const [addressErrors, setAddressErrors] = useState<FieldErrors>({});

  async function handleSubmit() {
    if (submitting) return;

    setClientErrors({});
    setAddressErrors({});
    setSubmitting(true);

    try {
      const createRes = await createClient(client);
      if (!createRes.ok || !createRes.body.success) {
        if (createRes.status === 422) {
          setClientErrors(fieldErrorsFromBody(createRes.body));
        } else {
          toast.error(
            (createRes.body.message as string) ??
              "Impossible de créer le client.",
          );
        }
        return;
      }

      const clientId = createRes.body.client_id as string;

      if (address.country_id != null && address.street && address.city) {
        const addrRes = await createAddress(
          { type: "client", clientId },
          {
            name: address.name || "Adresse principale",
            street: address.street,
            additional_street: address.additional_street,
            city: address.city,
            zip_code: address.zip_code,
            country_id: address.country_id,
          },
        );
        if (!addrRes.ok || !addrRes.body.success) {
          if (addrRes.status === 422) {
            setAddressErrors(fieldErrorsFromBody(addrRes.body));
          } else {
            toast.error(
              (addrRes.body.message as string) ??
                "Client créé, mais l'adresse n'a pas pu être enregistrée. Vous pouvez l'ajouter depuis la fiche client.",
            );
          }
          router.push(`/clients/${clientId}`);
          return;
        }
      }

      toast.success("Client créé.");
      router.push(`/clients/${clientId}`);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>Créer un compte client</CardTitle>
        <CardDescription>
          Créez un compte de base avec les informations principales.
        </CardDescription>
      </CardHeader>

      <CardContent>
        <ClientForm
          client={client}
          address={address}
          onClientChange={setClient}
          onAddressChange={setAddress}
          fieldErrors={clientErrors}
          addressErrors={addressErrors}
        />
      </CardContent>

      <CardFooter className="justify-end gap-2">
        <Button
          variant="outline"
          type="button"
          onClick={() => router.push("/clients")}
        >
          Annuler
        </Button>
        <Button type="button" onClick={handleSubmit} disabled={submitting}>
          {submitting ? "Création…" : "Créer le compte client"}
        </Button>
      </CardFooter>
    </Card>
  );
}
