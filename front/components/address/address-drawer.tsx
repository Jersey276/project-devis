"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
} from "@/components/ui/drawer";
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

type AddressDrawerProps = {
  ownerType: "user" | "client";
  ownerId: string;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  address?: ExistingAddress | null;
  onSaved: () => void;
};

const FORM_ID = "address-form";

export default function AddressDrawer({
  ownerType,
  ownerId,
  open,
  onOpenChange,
  address,
  onSaved,
}: AddressDrawerProps) {
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
        toast.success(isEdit ? "Adresse mise à jour." : "Adresse ajoutée.");
        onSaved();
        onOpenChange(false);
        return;
      }
      if (status === 422 && Array.isArray(body.field_errors)) {
        setFieldErrors(fieldErrorsFromBody(body));
        return;
      }
      toast.error(body.message ?? "Une erreur est survenue.");
    } catch {
      toast.error("Une erreur est survenue.");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <Drawer open={open} onOpenChange={onOpenChange} direction="right">
      <DrawerContent>
        <DrawerHeader>
          <DrawerTitle>
            {address ? "Modifier l'adresse" : "Nouvelle adresse"}
          </DrawerTitle>
        </DrawerHeader>
        <div className="flex-1 overflow-y-auto p-4">
          <AddressForm
            key={address?.id ?? "new"}
            formId={FORM_ID}
            initialValues={address ?? undefined}
            fieldErrors={fieldErrors}
            onSubmit={handleSubmit}
          />
        </div>
        <DrawerFooter>
          <Button type="submit" form={FORM_ID} disabled={submitting}>
            {submitting ? "Enregistrement…" : "Enregistrer"}
          </Button>
          <DrawerClose asChild>
            <Button type="button" variant="outline">
              Annuler
            </Button>
          </DrawerClose>
        </DrawerFooter>
      </DrawerContent>
    </Drawer>
  );
}
