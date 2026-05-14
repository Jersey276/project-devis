"use client";

import { useTranslations } from "next-intl";
import {
  Field,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { toErrorProps } from "@/lib/api";
import type { BackendAddress, BackendClient } from "@/types/backend";

type QuoteStepBasicInfoProps = {
  projectName: string;
  clientId: string;
  addressId: number | null;
  isReadonly: boolean;
  clients: BackendClient[];
  addresses: BackendAddress[];
  nameErrors?: string[];
  clientErrors?: string[];
  addressErrors?: string[];
  onProjectNameChange: (value: string) => void;
  onClientIdChange: (value: string) => void;
  onAddressIdChange: (value: number | null) => void;
};

export default function QuoteStepBasicInfo({
  projectName,
  clientId,
  addressId,
  isReadonly,
  clients,
  addresses,
  nameErrors,
  clientErrors,
  addressErrors,
  onProjectNameChange,
  onClientIdChange,
  onAddressIdChange,
}: QuoteStepBasicInfoProps) {
  const t = useTranslations("quote.steps.basicInfo");
  const hasNameError = !!nameErrors?.length;
  const hasClientError = !!clientErrors?.length;
  const hasAddressError = !!addressErrors?.length;

  const selectedClient = clientId
    ? clients.find((c) => c.client_id === clientId) ?? null
    : null;
  const selectedAddress =
    addressId != null
      ? addresses.find((a) => a.id === addressId) ?? null
      : null;

  const clientPlaceholder =
    clients.length === 0
      ? t("clientPlaceholderEmpty")
      : t("clientPlaceholder");

  const addressPlaceholder = !clientId
    ? t("addressPlaceholderNoClient")
    : addresses.length === 0
      ? t("addressPlaceholderEmpty")
      : t("addressPlaceholder");

  return (
    <div className="grid gap-4 md:max-w-xl">
      <Field data-invalid={hasNameError}>
        <FieldLabel htmlFor="project-name">{t("nameLabel")}</FieldLabel>
        <Input
          id="project-name"
          name="name"
          value={projectName}
          onChange={(event) => onProjectNameChange(event.target.value)}
          placeholder={t("namePlaceholder")}
          disabled={isReadonly}
          aria-invalid={hasNameError}
        />
        <FieldError errors={toErrorProps(nameErrors)} />
      </Field>

      <Field data-invalid={hasClientError}>
        <FieldLabel htmlFor="client-picker">{t("clientLabel")}</FieldLabel>
        <Combobox
          items={clients}
          value={selectedClient}
          onValueChange={(client: BackendClient | null) =>
            onClientIdChange(client ? client.client_id : "")
          }
          itemToStringLabel={(client: BackendClient) =>
            `${client.first_name} ${client.last_name}${client.company ? ` — ${client.company}` : ""}`
          }
        >
          <ComboboxInput
            id="client-picker"
            name="client_id"
            placeholder={clientPlaceholder}
            disabled={isReadonly || clients.length === 0}
            aria-invalid={hasClientError}
          />
          <ComboboxContent>
            <ComboboxEmpty>{t("clientEmpty")}</ComboboxEmpty>
            <ComboboxList>
              {(client: BackendClient) => (
                <ComboboxItem key={client.client_id} value={client}>
                  {client.first_name} {client.last_name}
                  {client.company ? ` — ${client.company}` : ""}
                </ComboboxItem>
              )}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
        <FieldError errors={toErrorProps(clientErrors)} />
      </Field>

      <Field data-invalid={hasAddressError}>
        <FieldLabel htmlFor="address-picker">{t("addressLabel")}</FieldLabel>
        <Combobox
          items={addresses}
          value={selectedAddress}
          onValueChange={(address: BackendAddress | null) =>
            onAddressIdChange(address ? address.id : null)
          }
          itemToStringLabel={(address: BackendAddress) =>
            `${address.name} — ${address.street}, ${address.zip_code} ${address.city}`
          }
        >
          <ComboboxInput
            id="address-picker"
            name="address_id"
            placeholder={addressPlaceholder}
            disabled={isReadonly || !clientId || addresses.length === 0}
            aria-invalid={hasAddressError}
          />
          <ComboboxContent>
            <ComboboxEmpty>{t("addressEmpty")}</ComboboxEmpty>
            <ComboboxList>
              {(address: BackendAddress) => (
                <ComboboxItem key={address.id} value={address}>
                  <div className="flex flex-col">
                    <span className="font-medium">{address.name}</span>
                    <span className="text-xs text-muted-foreground">
                      {address.street}, {address.zip_code} {address.city}
                    </span>
                  </div>
                </ComboboxItem>
              )}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
        <FieldError errors={toErrorProps(addressErrors)} />
      </Field>
    </div>
  );
}
