"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogContent,
  ResponsiveDialogFooter,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/custom/responsive-dialog";
import ClientForm, {
  type ClientFormValues,
  EMPTY_CLIENT_VALUES,
} from "@/components/user/client/client-form";
import {
  EMPTY_ADDRESS_VALUES,
  type AddressValues,
} from "@/components/address/address-form";
import { fieldErrorsFromBody, type FieldErrors } from "@/lib/api";
import { createClient } from "@/lib/services/clients";
import { createAddress } from "@/lib/services/addresses";
import { toast } from "sonner";

type ClientDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSaved: () => void;
};

export default function ClientDialog({
  open,
  onOpenChange,
  onSaved,
}: ClientDialogProps) {
  const t = useTranslations("client.dialog");
  const tCreate = useTranslations("client.create");
  const tCommon = useTranslations("common");
  const [client, setClient] = useState<ClientFormValues>(EMPTY_CLIENT_VALUES);
  const [address, setAddress] = useState<AddressValues>(EMPTY_ADDRESS_VALUES);
  const [clientErrors, setClientErrors] = useState<FieldErrors>({});
  const [addressErrors, setAddressErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setClient(EMPTY_CLIENT_VALUES);
      setAddress(EMPTY_ADDRESS_VALUES);
      setClientErrors({});
      setAddressErrors({});
      setSubmitting(false);
    }
  }, [open]);

  async function handleSave() {
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
          toast.error(createRes.body.message ?? tCreate("createFailedToast"));
        }
        return;
      }

      const clientId = createRes.body.client_id as string;

      if (address.country_id != null && address.street && address.city) {
        const addrRes = await createAddress(
          { type: "client", clientId },
          {
            name: address.name || tCreate("addressDefaultName"),
            street: address.street,
            additional_street: address.additional_street,
            city: address.city,
            zip_code: address.zip_code,
            country_id: address.country_id,
          },
        );
        if (!addrRes.ok || !addrRes.body.success) {
          toast.error(addrRes.body.message ?? tCreate("addressFailedToast"));
        }
      }

      toast.success(t("createSuccessToast"));
      onSaved();
      onOpenChange(false);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent>
        <ResponsiveDialogHeader>
          <ResponsiveDialogTitle>{t("createTitle")}</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        <ResponsiveDialogBody>
          <ClientForm
            client={client}
            onClientChange={setClient}
            fieldErrors={clientErrors}
            address={address}
            onAddressChange={setAddress}
            addressErrors={addressErrors}
          />
        </ResponsiveDialogBody>
        <ResponsiveDialogFooter>
          <Button type="button" onClick={handleSave} disabled={submitting}>
            {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
          </Button>
          <Button
            type="button"
            variant="outline"
            onClick={() => onOpenChange(false)}
          >
            {tCommon("actions.cancel")}
          </Button>
        </ResponsiveDialogFooter>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
