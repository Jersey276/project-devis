"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
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
  const t = useTranslations("client.create");
  const tCommon = useTranslations("common");
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
        const parsed = fieldErrorsFromBody(createRes.body);
        if (Object.keys(parsed).length > 0) {
          setClientErrors(parsed);
        } else {
          toast.error(
            (createRes.body.message as string) ?? t("createFailedToast"),
          );
        }
        return;
      }

      const clientId = createRes.body.client_id as string;

      if (address.country_id != null && address.street && address.city) {
        const addrRes = await createAddress(
          { type: "client", clientId },
          {
            name: address.name || t("addressDefaultName"),
            street: address.street,
            additional_street: address.additional_street,
            city: address.city,
            zip_code: address.zip_code,
            country_id: address.country_id,
          },
        );
        if (!addrRes.ok || !addrRes.body.success) {
          const parsedAddr = fieldErrorsFromBody(addrRes.body);
          if (Object.keys(parsedAddr).length > 0) {
            setAddressErrors(parsedAddr);
          } else {
            toast.error(
              (addrRes.body.message as string) ?? t("addressFailedToast"),
            );
          }
          router.push(`/clients/${clientId}`);
          return;
        }
      }

      toast.success(t("successToast"));
      router.push(`/clients/${clientId}`);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Card className="max-w-3xl">
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
        <CardDescription>{t("description")}</CardDescription>
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
          {tCommon("actions.cancel")}
        </Button>
        <Button type="button" onClick={handleSubmit} disabled={submitting}>
          {submitting ? t("submitting") : t("submit")}
        </Button>
      </CardFooter>
    </Card>
  );
}
