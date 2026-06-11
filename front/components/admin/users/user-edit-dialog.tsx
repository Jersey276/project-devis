"use client";

import { useEffect, useState } from "react";
import { useDialogSubmit } from "@/hooks/use-dialog-submit";
import { useTranslations } from "next-intl";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Field,
  FieldError,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toErrorProps } from "@/lib/api";
import { updateAdminUser } from "@/lib/services/admin-users";
import {
  type AdminUserAccount,
  type AdminUserRole,
} from "@/components/admin/types";

type UserEditDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  user: AdminUserAccount | null;
  onSaved: () => void;
};

const FORM_ID = "admin-user-edit-form";

export default function UserEditDialog({
  open,
  onOpenChange,
  user,
  onSaved,
}: UserEditDialogProps) {
  const t = useTranslations("admin.users.dialog");
  const tCommon = useTranslations("common");

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [phone, setPhone] = useState("");
  const [company, setCompany] = useState("");
  const [siren, setSiren] = useState("");
  const [vat, setVat] = useState("");
  const [role, setRole] = useState<AdminUserRole>("user");
  const [plan, setPlan] = useState("");
  const { fieldErrors, setFieldErrors, submitting, submit } = useDialogSubmit(
    tCommon("errors.generic"),
  );

  useEffect(() => {
    setFirstName(user?.first_name ?? "");
    setLastName(user?.last_name ?? "");
    setPhone(user?.phone ?? "");
    setCompany(user?.company ?? "");
    setSiren(user?.siren ?? "");
    setVat(user?.vat ?? "");
    setRole(user?.role ?? "user");
    setPlan(user?.plan ?? "");
    setFieldErrors({});
  }, [user, setFieldErrors]);

  async function handleSubmit(event: React.FormEvent) {
    event.preventDefault();
    if (!user) return;
    await submit({
      request: () =>
        updateAdminUser(user.user_id, {
          first_name: firstName,
          last_name: lastName,
          email: user.email,
          role,
          plan,
          phone,
          company,
          siren,
          vat,
        }),
      successMessage: t("updateSuccessToast"),
      onSuccess: onSaved,
      onClose: onOpenChange,
    });
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="p-6 sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("title")}</DialogTitle>
        </DialogHeader>

        <form
          id={FORM_ID}
          className="grid gap-4"
          onSubmit={handleSubmit}
          noValidate
        >
          <FieldGroup>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field data-invalid={!!fieldErrors.first_name?.length}>
                <FieldLabel htmlFor="admin_user_first_name">
                  {t("firstNameLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_first_name"
                  name="first_name"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  aria-invalid={!!fieldErrors.first_name?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.first_name)} />
              </Field>

              <Field data-invalid={!!fieldErrors.last_name?.length}>
                <FieldLabel htmlFor="admin_user_last_name">
                  {t("lastNameLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_last_name"
                  name="last_name"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  aria-invalid={!!fieldErrors.last_name?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.last_name)} />
              </Field>
            </div>

            <Field>
              <FieldLabel htmlFor="admin_user_email">
                {t("emailLabel")}
              </FieldLabel>
              <Input
                id="admin_user_email"
                name="email"
                type="email"
                value={user?.email ?? ""}
                readOnly
                aria-readonly="true"
              />
            </Field>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field data-invalid={!!fieldErrors.role?.length}>
                <FieldLabel htmlFor="admin_user_role">
                  {t("roleLabel")}
                </FieldLabel>
                <Select
                  value={role}
                  onValueChange={(value) => setRole(value as AdminUserRole)}
                >
                  <SelectTrigger
                    id="admin_user_role"
                    className="w-full"
                    aria-invalid={!!fieldErrors.role?.length}
                  >
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="user">{t("roleUser")}</SelectItem>
                    <SelectItem value="admin">{t("roleAdmin")}</SelectItem>
                  </SelectContent>
                </Select>
                <FieldError errors={toErrorProps(fieldErrors.role)} />
              </Field>

              <Field data-invalid={!!fieldErrors.plan?.length}>
                <FieldLabel htmlFor="admin_user_plan">
                  {t("planLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_plan"
                  name="plan"
                  value={plan}
                  onChange={(e) => setPlan(e.target.value)}
                  placeholder={t("planPlaceholder")}
                  aria-invalid={!!fieldErrors.plan?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.plan)} />
              </Field>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field data-invalid={!!fieldErrors.phone?.length}>
                <FieldLabel htmlFor="admin_user_phone">
                  {t("phoneLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_phone"
                  name="phone"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  aria-invalid={!!fieldErrors.phone?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.phone)} />
              </Field>

              <Field data-invalid={!!fieldErrors.company?.length}>
                <FieldLabel htmlFor="admin_user_company">
                  {t("companyLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_company"
                  name="company"
                  value={company}
                  onChange={(e) => setCompany(e.target.value)}
                  aria-invalid={!!fieldErrors.company?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.company)} />
              </Field>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
              <Field data-invalid={!!fieldErrors.siren?.length}>
                <FieldLabel htmlFor="admin_user_siren">
                  {t("sirenLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_siren"
                  name="siren"
                  value={siren}
                  onChange={(e) => setSiren(e.target.value)}
                  aria-invalid={!!fieldErrors.siren?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.siren)} />
              </Field>

              <Field data-invalid={!!fieldErrors.vat?.length}>
                <FieldLabel htmlFor="admin_user_vat">
                  {t("vatLabel")}
                </FieldLabel>
                <Input
                  id="admin_user_vat"
                  name="vat"
                  value={vat}
                  onChange={(e) => setVat(e.target.value)}
                  aria-invalid={!!fieldErrors.vat?.length}
                />
                <FieldError errors={toErrorProps(fieldErrors.vat)} />
              </Field>
            </div>
          </FieldGroup>
        </form>

        <DialogFooter>
          <DialogClose asChild>
            <Button type="button" variant="outline">
              {tCommon("actions.cancel")}
            </Button>
          </DialogClose>
          <Button type="submit" form={FORM_ID} disabled={submitting || !user}>
            {submitting ? tCommon("actions.saving") : tCommon("actions.save")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
