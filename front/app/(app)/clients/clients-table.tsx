"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import {
  DataTable,
  DataTableBodyRows,
  DataTableCell,
  DataTableHeader,
  DataTableHead,
  DataTableRow,
  DataTableRowActions,
  DataTableSortableHead,
} from "@/components/custom/data-table";
import type { DataTableRowAction } from "@/components/custom/data-table";
import { EyeIcon, LinkIcon, MailIcon, TrashIcon } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { toast } from "sonner";
import { archiveClient, sendClientInvitation } from "@/lib/services/clients";
import type { BackendClient } from "@/types/backend";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";

type ClientsTableProps = {
  data: BackendClient[];
  onArchived?: () => void;
};

export function ClientsTable({ data, onArchived }: ClientsTableProps) {
  const t = useTranslations("client.list");
  const tCommon = useTranslations("common");
  const tInvite = useTranslations("auth.invite");

  const [pendingInvite, setPendingInvite] = useState<BackendClient | null>(null);
  const [inviting, setInviting] = useState(false);

  // Row actions are computed per-row via the callback pattern — but DataTable
  // uses a shared array. We hide "Inviter" globally and rely on the callback
  // to guard against linked clients; the hidden flag on the action prevents
  // it from appearing when the row has no email.
  // For the "linked" case we use disabled + a guard in the callback.
  const row_actions: DataTableRowAction[] = [
    {
      type: "link",
      label: t("actions.view"),
      href: "/clients/{id}",
      icon: EyeIcon,
    },
    {
      type: "callback",
      label: tInvite("sendButton"),
      icon: MailIcon,
      callback: (row) => {
        const client = row as BackendClient;
        if (client.linked_user_id) return; // already linked — guard
        if (!client.email) {
          toast.error(tInvite("noEmail"));
          return;
        }
        setPendingInvite(client);
      },
    },
    {
      type: "callback",
      label: tCommon("actions.delete"),
      icon: TrashIcon,
      callback: async (row) => {
        const client = row as BackendClient;
        const { ok, body } = await archiveClient(client.client_id);
        if (ok && body.success) {
          toast.success(t("deleteSuccessToast"));
          onArchived?.();
        } else {
          toast.error((body.message as string) ?? t("deleteFailedToast"));
        }
      },
    },
  ];

  async function handleConfirmInvite() {
    if (!pendingInvite) return;
    setInviting(true);
    try {
      const { ok, body } = await sendClientInvitation(pendingInvite.client_id);
      if (ok && body.success) {
        toast.success(tInvite("sendSuccess"));
      } else {
        toast.error((body.message as string) ?? tInvite("sendError"));
      }
    } catch {
      toast.error(tInvite("sendError"));
    } finally {
      setInviting(false);
      setPendingInvite(null);
    }
  }

  return (
    <>
      <DataTable
        datas={data}
        sortBy="client_id"
        sortDirection="asc"
        row_actions={row_actions}
      >
        <DataTableHeader>
          <DataTableRow>
            <DataTableSortableHead name="first_name">
              {t("columns.firstName")}
            </DataTableSortableHead>
            <DataTableSortableHead name="last_name">
              {t("columns.lastName")}
            </DataTableSortableHead>
            <DataTableSortableHead name="email">
              {t("columns.email")}
            </DataTableSortableHead>
            <DataTableSortableHead name="company">
              {t("columns.company")}
            </DataTableSortableHead>
            <DataTableHead>{t("columns.actions")}</DataTableHead>
          </DataTableRow>
        </DataTableHeader>
        <DataTableBodyRows<BackendClient>
          render={(client) => (
            <DataTableRow key={client.client_id}>
              <DataTableCell>{client.first_name}</DataTableCell>
              <DataTableCell>{client.last_name}</DataTableCell>
              <DataTableCell>{client.email}</DataTableCell>
              <DataTableCell>
                <span className="flex items-center gap-2">
                  {client.company}
                  {client.linked_user_id && (
                    <Badge variant="secondary" className="gap-1 text-xs">
                      <LinkIcon className="size-3" />
                      {t("linked")}
                    </Badge>
                  )}
                </span>
              </DataTableCell>
              <DataTableCell>
                <DataTableRowActions id={client.client_id} row={client} />
              </DataTableCell>
            </DataTableRow>
          )}
        />
      </DataTable>

      <AlertDialog
        open={!!pendingInvite}
        onOpenChange={(open) => { if (!open) setPendingInvite(null); }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{tInvite("sendConfirmTitle")}</AlertDialogTitle>
            <AlertDialogDescription>
              {tInvite("sendConfirmDescription", { email: pendingInvite?.email ?? "" })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={inviting}>
              {tCommon("actions.cancel")}
            </AlertDialogCancel>
            <AlertDialogAction onClick={handleConfirmInvite} disabled={inviting}>
              {inviting ? tCommon("actions.saving") : tInvite("sendConfirm")}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
