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
import {
  apiFetch,
  fieldErrorsFromBody,
  FieldErrors,
} from "@/lib/api";
import { toast } from "sonner";

export type ExistingAddress = AddressValues & { id: number };

type AddressDrawerProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  address?: ExistingAddress | null;
  onSaved: () => void;
};

const FORM_ID = "address-form";

export default function AddressDrawer({
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
    const path = isEdit
      ? `/api/users/me/addresses/${address!.id}`
      : "/api/users/me/addresses";
    try {
      const { ok, status, body } = await apiFetch(path, {
        method: isEdit ? "PUT" : "POST",
        body: JSON.stringify(values),
      });
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
