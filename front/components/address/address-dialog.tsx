"use client";

import { useEffect } from "react";
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
import AddressForm, {
  type AddressValues,
} from "@/components/address/address-form";
import {
  buildOwner,
  createAddress,
  updateAddress,
} from "@/lib/services/addresses";
import { useDialogSubmit } from "@/hooks/use-dialog-submit";

export type ExistingAddress = AddressValues & { id: number };

type AddressDialogProps = {
  ownerType: "user" | "client";
  ownerId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  address?: ExistingAddress | null;
  onSaved: () => void;
};

const FORM_ID = "address-form";

export default function AddressDialog({
  ownerType,
  ownerId,
  open,
  onOpenChange,
  address,
  onSaved,
}: AddressDialogProps) {
  const t = useTranslations("address.dialog");
  const tCommon = useTranslations("common");
  const { fieldErrors, setFieldErrors, submitting, submit } = useDialogSubmit(
    tCommon("errors.generic"),
  );

  useEffect(() => {
    if (!open) setFieldErrors({});
  }, [open, setFieldErrors]);

  async function handleSubmit(values: AddressValues) {
    if (values.country_id == null) {
      setFieldErrors({ country_id: [tCommon("validation.required")] });
      return;
    }
    const isEdit = address?.id != null;
    const owner = buildOwner(ownerType, ownerId);
    await submit({
      request: () =>
        isEdit
          ? updateAddress(owner, address!.id, values)
          : createAddress(owner, values),
      successMessage: isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent>
        <ResponsiveDialogHeader>
          <ResponsiveDialogTitle>
            {address ? t("editTitle") : t("createTitle")}
          </ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        <ResponsiveDialogBody>
          <AddressForm
            key={address?.id ?? "new"}
            formId={FORM_ID}
            initialValues={address ?? undefined}
            fieldErrors={fieldErrors}
            onSubmit={handleSubmit}
          />
        </ResponsiveDialogBody>
        <ResponsiveDialogFooter>
          <Button type="submit" form={FORM_ID} disabled={submitting}>
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
