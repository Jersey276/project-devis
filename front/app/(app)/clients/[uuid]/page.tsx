"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { useTranslations } from "next-intl";
import { toast } from "sonner";
import { LinkIcon, PencilIcon, Trash2Icon } from "lucide-react";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import AddressesTable from "@/components/address/addresses-table";
import { archiveClient, getClient } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";

export default function ClientProfilePage() {
  const { uuid } = useParams<{ uuid: string }>();
  const router = useRouter();
  const t = useTranslations("client.detail");
  const tCommon = useTranslations("common");
  const [client, setClient] = useState<BackendClient | null>(null);
  const [archiving, setArchiving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getClient(uuid).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setClient(body.client as BackendClient);
      } else {
        toast.error((body.message as string) ?? t("notFoundToast"));
      }
    });
    return () => {
      cancelled = true;
    };
  }, [uuid, t]);

  async function handleArchive() {
    if (archiving) return;
    setArchiving(true);
    try {
      const { ok, body } = await archiveClient(uuid);
      if (ok && body.success) {
        toast.success(t("deleteSuccessToast"));
        router.push("/clients");
      } else {
        toast.error((body.message as string) ?? t("deleteFailedToast"));
        setArchiving(false);
      }
    } catch {
      toast.error(t("deleteFailedToast"));
      setArchiving(false);
    }
  }

  if (!client) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          {tCommon("actions.loading")}
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="gap-4">
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>
              {t("idLabel")} {client.client_id} • {client.first_name}{" "}
              {client.last_name}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline">{t("badge")}</Badge>
            {client.linked_user_id && (
              <Badge variant="secondary" className="gap-1">
                <LinkIcon className="h-3 w-3" />
                {t("linkedBadge")}
              </Badge>
            )}
            {!client.linked_user_id && (
              <Button asChild size="sm" variant="outline">
                <Link
                  href={`/clients/${uuid}/edit`}
                  className="inline-flex items-center gap-1"
                >
                  <PencilIcon className="h-4 w-4" />
                  {tCommon("actions.edit")}
                </Link>
              </Button>
            )}
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  size="sm"
                  variant="destructive"
                  disabled={archiving}
                  className="inline-flex items-center gap-1"
                >
                  <Trash2Icon className="h-4 w-4" />
                  {tCommon("actions.delete")}
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>{t("deleteDialog.title")}</AlertDialogTitle>
                  <AlertDialogDescription>
                    {t("deleteDialog.description")}
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>
                    {tCommon("actions.cancel")}
                  </AlertDialogCancel>
                  <AlertDialogAction
                    variant="destructive"
                    onClick={handleArchive}
                  >
                    {tCommon("actions.delete")}
                  </AlertDialogAction>
                </AlertDialogFooter>
              </AlertDialogContent>
            </AlertDialog>
          </div>
        </div>
      </CardHeader>

      <CardContent className="grid gap-4">
        <Card size="sm" className="gap-3">
          <CardHeader className="pb-0">
            <CardTitle>{t("infoTitle")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1 text-sm">
            <p>
              <span className="font-medium">{t("info.firstName")}</span>{" "}
              {client.first_name}
            </p>
            <p>
              <span className="font-medium">{t("info.lastName")}</span>{" "}
              {client.last_name}
            </p>
            <p>
              <span className="font-medium">{t("info.email")}</span>{" "}
              {client.email}
            </p>
            {client.phone && (
              <p>
                <span className="font-medium">{t("info.phone")}</span>{" "}
                {client.phone}
              </p>
            )}
            {client.company && (
              <p>
                <span className="font-medium">{t("info.company")}</span>{" "}
                {client.company}
              </p>
            )}
            {client.siren && (
              <p>
                <span className="font-medium">{t("info.siren")}</span>{" "}
                {client.siren}
              </p>
            )}
            {client.siret && (
              <p>
                <span className="font-medium">{t("info.siret")}</span>{" "}
                {client.siret}
              </p>
            )}
            {client.vat && (
              <p>
                <span className="font-medium">{t("info.vat")}</span>{" "}
                {client.vat}
              </p>
            )}
          </CardContent>
        </Card>

        <div>
          <h3 className="mb-2 text-sm font-medium">{t("addressesTitle")}</h3>
          <AddressesTable ownerType="client" ownerId={uuid} />
        </div>
      </CardContent>
    </Card>
  );
}
