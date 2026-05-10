"use client";

import { Input } from "@/components/ui/input";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { FieldErrors, toErrorProps } from "@/lib/api";
import AddressForm, { type AddressValues } from "@/components/address/address-form";
import type { ClientPayload } from "@/lib/services/clients";

export type ClientFormValues = ClientPayload;

export const EMPTY_CLIENT_VALUES: ClientFormValues = {
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
  company: "",
  siren: "",
  vat: "",
};

type ClientFormProps = {
  formId?: string;
  client: ClientFormValues;
  onClientChange: (values: ClientFormValues) => void;
  fieldErrors?: FieldErrors;
  // Address fields are only relevant on create — the profile page manages
  // addresses separately via AddressesTable. Omit on edit.
  address?: AddressValues;
  onAddressChange?: (values: AddressValues) => void;
  addressErrors?: FieldErrors;
};

export default function ClientForm({
  formId,
  client,
  onClientChange,
  fieldErrors,
  address,
  onAddressChange,
  addressErrors,
}: ClientFormProps) {
  function update<K extends keyof ClientFormValues>(
    key: K,
    value: ClientFormValues[K],
  ) {
    onClientChange({ ...client, [key]: value });
  }

  return (
    <form id={formId ?? "create-client-form"} className="grid gap-6">
      <FieldGroup>
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors?.first_name?.length}>
            <FieldLabel htmlFor="first_name">Prénom</FieldLabel>
            <Input
              id="first_name"
              name="first_name"
              placeholder="Jean"
              value={client.first_name}
              onChange={(e) => update("first_name", e.target.value)}
              aria-invalid={!!fieldErrors?.first_name?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.first_name)} />
          </Field>

          <Field data-invalid={!!fieldErrors?.last_name?.length}>
            <FieldLabel htmlFor="last_name">Nom</FieldLabel>
            <Input
              id="last_name"
              name="last_name"
              placeholder="Dupont"
              value={client.last_name}
              onChange={(e) => update("last_name", e.target.value)}
              aria-invalid={!!fieldErrors?.last_name?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.last_name)} />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors?.email?.length}>
            <FieldLabel htmlFor="client_email">Adresse mail</FieldLabel>
            <Input
              id="client_email"
              name="email"
              type="email"
              placeholder="jean.dupont@email.com"
              value={client.email}
              onChange={(e) => update("email", e.target.value)}
              aria-invalid={!!fieldErrors?.email?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.email)} />
          </Field>

          <Field data-invalid={!!fieldErrors?.phone?.length}>
            <FieldLabel htmlFor="client_phone">Téléphone</FieldLabel>
            <Input
              id="client_phone"
              name="phone"
              placeholder="06 12 34 56 78"
              value={client.phone}
              onChange={(e) => update("phone", e.target.value)}
              aria-invalid={!!fieldErrors?.phone?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.phone)} />
          </Field>
        </div>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-3">
          <Field data-invalid={!!fieldErrors?.company?.length}>
            <FieldLabel htmlFor="client_company">Société</FieldLabel>
            <Input
              id="client_company"
              name="company"
              placeholder="Acme SARL"
              value={client.company}
              onChange={(e) => update("company", e.target.value)}
              aria-invalid={!!fieldErrors?.company?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.company)} />
          </Field>

          <Field data-invalid={!!fieldErrors?.siren?.length}>
            <FieldLabel htmlFor="client_siren">SIREN</FieldLabel>
            <Input
              id="client_siren"
              name="siren"
              placeholder="123 456 789"
              value={client.siren}
              onChange={(e) => update("siren", e.target.value)}
              aria-invalid={!!fieldErrors?.siren?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.siren)} />
          </Field>

          <Field data-invalid={!!fieldErrors?.vat?.length}>
            <FieldLabel htmlFor="client_vat">TVA</FieldLabel>
            <Input
              id="client_vat"
              name="vat"
              placeholder="FR12345678901"
              value={client.vat}
              onChange={(e) => update("vat", e.target.value)}
              aria-invalid={!!fieldErrors?.vat?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.vat)} />
          </Field>
        </div>
      </FieldGroup>

      {address && onAddressChange && (
        <div className="grid gap-3 rounded-lg border p-4">
          <h3 className="text-sm font-medium">Adresse principale</h3>
          <AddressForm
            initialValues={address}
            onChange={onAddressChange}
            fieldErrors={addressErrors}
          />
        </div>
      )}
    </form>
  );
}
