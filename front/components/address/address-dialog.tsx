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
import AddressForm, {
  type AddressValues,
} from "@/components/address/address-form";
import { fieldErrorsFromBody, FieldErrors } from "@/lib/api";
import {
  buildOwner,
  createAddress,
  updateAddress,
} from "@/lib/services/addresses";
import { toast } from "sonner";

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
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({});
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!open) {
      setFieldErrors({});
      setSubmitting(false);
    }
  }, [open]);

  async function handleSubmit(values: AddressValues) {
    setFieldErrors({});
    setSubmitting(true);
    const isEdit = address?.id != null;
    try {
      const owner = buildOwner(ownerType, ownerId);
      const { ok, status, body } = isEdit
        ? await updateAddress(owner, address!.id, values)
        : await createAddress(owner, values);
      if (ok && body.success) {
        toast.success(
          isEdit ? t("updateSuccessToast") : t("createSuccessToast"),
        );
        onSaved();
        onOpenChange(false);
        return;
      }
      if (status === 422 && Array.isArray(body.field_errors)) {
        setFieldErrors(fieldErrorsFromBody(body));
        return;
      }
      toast.error(body.message ?? tCommon("errors.generic"));
    } catch {
      toast.error(tCommon("errors.generic"));
    } finally {
      setSubmitting(false);
    }
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
