"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Input } from "@/components/ui/input";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { apiFetch, FieldErrors, toErrorProps } from "@/lib/api";

export type Country = {
  id: number;
  code: string;
  name: string;
};

export type AddressValues = {
  name: string;
  street: string;
  additional_street: string;
  city: string;
  zip_code: string;
  country_id: number | null;
};

export const EMPTY_ADDRESS_VALUES: AddressValues = {
  name: "",
  street: "",
  additional_street: "",
  city: "",
  zip_code: "",
  country_id: null,
};

type AddressFormProps = {
  formId?: string;
  initialValues?: AddressValues;
  fieldErrors?: FieldErrors;
  onSubmit?: (values: AddressValues) => void;
  onChange?: (values: AddressValues) => void;
};

export default function AddressForm({
  formId,
  initialValues,
  fieldErrors,
  onSubmit,
  onChange,
}: AddressFormProps) {
  const t = useTranslations("address.form");
  const [values, setValues] = useState<AddressValues>(
    () => initialValues ?? EMPTY_ADDRESS_VALUES,
  );
  const [countries, setCountries] = useState<Country[]>([]);

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/users/countries").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && Array.isArray(body.countries)) {
        setCountries(body.countries as Country[]);
      }
    });
    return () => {
      cancelled = true;
    };
  }, []);

  function update<K extends keyof AddressValues>(key: K, value: AddressValues[K]) {
    setValues((prev) => {
      const next = { ...prev, [key]: value };
      onChange?.(next);
      return next;
    });
  }

  const fields = (
    <FieldGroup>
        <Field data-invalid={!!fieldErrors?.name?.length}>
          <FieldLabel htmlFor="address_name">{t("nameLabel")}</FieldLabel>
          <Input
            id="address_name"
            name="name"
            placeholder={t("namePlaceholder")}
            value={values.name}
            onChange={(e) => update("name", e.target.value)}
            aria-invalid={!!fieldErrors?.name?.length}
          />
          <FieldError errors={toErrorProps(fieldErrors?.name)} />
        </Field>

        <Field data-invalid={!!fieldErrors?.street?.length}>
          <FieldLabel htmlFor="address_street">{t("streetLabel")}</FieldLabel>
          <Input
            id="address_street"
            name="street"
            placeholder={t("streetPlaceholder")}
            value={values.street}
            onChange={(e) => update("street", e.target.value)}
            aria-invalid={!!fieldErrors?.street?.length}
          />
          <FieldError errors={toErrorProps(fieldErrors?.street)} />
        </Field>

        <Field data-invalid={!!fieldErrors?.additional_street?.length}>
          <FieldLabel htmlFor="address_additional_street">{t("additionalStreetLabel")}</FieldLabel>
          <Input
            id="address_additional_street"
            name="additional_street"
            placeholder={t("additionalStreetPlaceholder")}
            value={values.additional_street}
            onChange={(e) => update("additional_street", e.target.value)}
            aria-invalid={!!fieldErrors?.additional_street?.length}
          />
          <FieldError errors={toErrorProps(fieldErrors?.additional_street)} />
        </Field>

        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          <Field data-invalid={!!fieldErrors?.city?.length}>
            <FieldLabel htmlFor="address_city">{t("cityLabel")}</FieldLabel>
            <Input
              id="address_city"
              name="city"
              placeholder={t("cityPlaceholder")}
              value={values.city}
              onChange={(e) => update("city", e.target.value)}
              aria-invalid={!!fieldErrors?.city?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.city)} />
          </Field>

          <Field data-invalid={!!fieldErrors?.zip_code?.length}>
            <FieldLabel htmlFor="address_zip_code">{t("zipCodeLabel")}</FieldLabel>
            <Input
              id="address_zip_code"
              name="zip_code"
              placeholder={t("zipCodePlaceholder")}
              value={values.zip_code}
              onChange={(e) => update("zip_code", e.target.value)}
              aria-invalid={!!fieldErrors?.zip_code?.length}
            />
            <FieldError errors={toErrorProps(fieldErrors?.zip_code)} />
          </Field>
        </div>

        <Field data-invalid={!!fieldErrors?.country_id?.length}>
          <FieldLabel htmlFor="address_country">{t("countryLabel")}</FieldLabel>
          <Combobox
            items={countries}
            value={
              values.country_id != null
                ? countries.find((c) => c.id === values.country_id) ?? null
                : null
            }
            onValueChange={(item: Country | null) =>
              update("country_id", item ? item.id : null)
            }
            itemToStringLabel={(item: Country) => item.name}
          >
            <ComboboxInput
              id="address_country"
              name="country_id"
              placeholder={t("countryPlaceholder")}
              aria-invalid={!!fieldErrors?.country_id?.length}
            />
            <ComboboxContent>
              <ComboboxEmpty>{t("countryEmpty")}</ComboboxEmpty>
              <ComboboxList>
                {(country: Country) => (
                  <ComboboxItem key={country.id} value={country}>
                    {country.name}
                  </ComboboxItem>
                )}
              </ComboboxList>
            </ComboboxContent>
          </Combobox>
          <FieldError errors={toErrorProps(fieldErrors?.country_id)} />
        </Field>
      </FieldGroup>
  );

  if (!onSubmit) {
    return <div className="grid gap-4">{fields}</div>;
  }

  return (
    <form
      id={formId ?? "address-form"}
      className="grid gap-4"
      onSubmit={(e) => {
        e.preventDefault();
        onSubmit(values);
      }}
      noValidate
    >
      {fields}
    </form>
  );
}
