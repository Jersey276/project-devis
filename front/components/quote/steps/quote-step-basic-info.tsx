"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { PlusIcon } from "lucide-react";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
  ComboboxSeparator,
} from "@/components/ui/combobox";
import { toErrorProps } from "@/lib/api";
import type { BackendAddress, BackendClient } from "@/types/backend";
import AddressDialog from "@/components/address/address-dialog";
import ClientDialog from "@/components/user/client/client-dialog";

type QuoteStepBasicInfoProps = {
  projectName: string;
  clientId: string;
  addressId: number | null;
  userAddressId: number | null;
  isReadonly: boolean;
  clients: BackendClient[];
  addresses: BackendAddress[];
  userAddresses: BackendAddress[];
  userId: string;
  nameErrors?: string[];
  clientErrors?: string[];
  addressErrors?: string[];
  userAddressErrors?: string[];
  onProjectNameChange: (value: string) => void;
  onClientIdChange: (value: string) => void;
  onAddressIdChange: (value: number | null) => void;
  onUserAddressIdChange: (value: number | null) => void;
  onClientCreated: () => void;
  onUserAddressCreated: () => void;
  onClientAddressCreated: () => void;
};

export default function QuoteStepBasicInfo({
  projectName,
  clientId,
  addressId,
  userAddressId,
  isReadonly,
  clients,
  addresses,
  userAddresses,
  userId,
  nameErrors,
  clientErrors,
  addressErrors,
  userAddressErrors,
  onProjectNameChange,
  onClientIdChange,
  onAddressIdChange,
  onUserAddressIdChange,
  onClientCreated,
  onUserAddressCreated,
  onClientAddressCreated,
}: QuoteStepBasicInfoProps) {
  const t = useTranslations("quote.steps.basicInfo");
  const [addUserAddressOpen, setAddUserAddressOpen] = useState(false);
  const [addClientOpen, setAddClientOpen] = useState(false);
  const [addClientAddressOpen, setAddClientAddressOpen] = useState(false);

  const hasNameError = !!nameErrors?.length;
  const hasClientError = !!clientErrors?.length;
  const hasAddressError = !!addressErrors?.length;
  const hasUserAddressError = !!userAddressErrors?.length;

  const selectedClient = clientId
    ? (clients.find((c) => c.client_id === clientId) ?? null)
    : null;
  const selectedAddress =
    addressId != null
      ? (addresses.find((a) => a.id === addressId) ?? null)
      : null;
  const selectedUserAddress =
    userAddressId != null
      ? (userAddresses.find((a) => a.id === userAddressId) ?? null)
      : null;

  const clientPlaceholder =
    clients.length === 0 ? t("clientPlaceholderEmpty") : t("clientPlaceholder");

  const addressPlaceholder = !clientId
    ? t("addressPlaceholderNoClient")
    : addresses.length === 0
      ? t("addressPlaceholderEmpty")
      : t("addressPlaceholder");

  const userAddressPlaceholder =
    userAddresses.length === 0
      ? t("userAddressPlaceholderEmpty")
      : t("userAddressPlaceholder");

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

      <Field data-invalid={hasUserAddressError}>
        <FieldLabel htmlFor="user-address-picker">
          {t("userAddressLabel")}
        </FieldLabel>
        <Combobox
          items={userAddresses}
          value={selectedUserAddress}
          onValueChange={(address: BackendAddress | null) =>
            onUserAddressIdChange(address ? address.id : null)
          }
          itemToStringLabel={(address: BackendAddress) =>
            `${address.name} — ${address.street}, ${address.zip_code} ${address.city}`
          }
        >
          <ComboboxInput
            id="user-address-picker"
            name="user_address_id"
            placeholder={userAddressPlaceholder}
            disabled={isReadonly}
            aria-invalid={hasUserAddressError}
          />
          <ComboboxContent>
            {!isReadonly && !!userId && (
              <>
                <div className="p-1">
                  <button
                    type="button"
                    className="flex w-full items-center gap-1.5 rounded-sm px-2 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => setAddUserAddressOpen(true)}
                  >
                    <PlusIcon className="size-3.5" />
                    {t("addUserAddressButton")}
                  </button>
                </div>
                <ComboboxSeparator />
              </>
            )}
            <ComboboxEmpty>{t("userAddressEmpty")}</ComboboxEmpty>
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
        <FieldError errors={toErrorProps(userAddressErrors)} />
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
            disabled={isReadonly}
            aria-invalid={hasClientError}
          />
          <ComboboxContent>
            {!isReadonly && (
              <>
                <div className="p-1">
                  <button
                    type="button"
                    className="flex w-full items-center gap-1.5 rounded-sm px-2 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => setAddClientOpen(true)}
                  >
                    <PlusIcon className="size-3.5" />
                    {t("addClientButton")}
                  </button>
                </div>
                <ComboboxSeparator />
              </>
            )}
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
            disabled={isReadonly || !clientId}
            aria-invalid={hasAddressError}
          />
          <ComboboxContent>
            {!isReadonly && !!clientId && (
              <>
                <div className="p-1">
                  <button
                    type="button"
                    className="flex w-full items-center gap-1.5 rounded-sm px-2 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                    onMouseDown={(e) => e.preventDefault()}
                    onClick={() => setAddClientAddressOpen(true)}
                  >
                    <PlusIcon className="size-3.5" />
                    {t("addAddressButton")}
                  </button>
                </div>
                <ComboboxSeparator />
              </>
            )}
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

      <AddressDialog
        ownerType="user"
        ownerId={userId}
        open={addUserAddressOpen}
        onOpenChange={setAddUserAddressOpen}
        onSaved={onUserAddressCreated}
      />

      <ClientDialog
        open={addClientOpen}
        onOpenChange={setAddClientOpen}
        onSaved={onClientCreated}
      />

      {clientId && (
        <AddressDialog
          ownerType="client"
          ownerId={clientId}
          open={addClientAddressOpen}
          onOpenChange={setAddClientAddressOpen}
          onSaved={onClientAddressCreated}
        />
      )}
    </div>
  );
}
