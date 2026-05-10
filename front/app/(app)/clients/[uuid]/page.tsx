"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { PencilIcon, Trash2Icon } from "lucide-react";
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
  const [client, setClient] = useState<BackendClient | null>(null);
  const [archiving, setArchiving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getClient(uuid).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setClient(body.client as BackendClient);
      } else {
        toast.error((body.message as string) ?? "Client introuvable.");
      }
    });
    return () => {
      cancelled = true;
    };
  }, [uuid]);

  async function handleArchive() {
    if (archiving) return;
    setArchiving(true);
    try {
      const { ok, body } = await archiveClient(uuid);
      if (ok && body.success) {
        toast.success("Client supprimé.");
        router.push("/clients");
      } else {
        toast.error(
          (body.message as string) ?? "Impossible de supprimer le client.",
        );
        setArchiving(false);
      }
    } catch {
      toast.error("Impossible de supprimer le client.");
      setArchiving(false);
    }
  }

  if (!client) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          Chargement…
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="gap-4">
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            <CardTitle>Profil client</CardTitle>
            <CardDescription>
              ID: {client.client_id} • {client.first_name} {client.last_name}
            </CardDescription>
          </div>
          <div className="flex items-center gap-2">
            <Badge variant="outline">Client</Badge>
            <Button asChild size="sm" variant="outline">
              <Link
                href={`/clients/${uuid}/edit`}
                className="inline-flex items-center gap-1"
              >
                <PencilIcon className="h-4 w-4" />
                Modifier
              </Link>
            </Button>
            <AlertDialog>
              <AlertDialogTrigger asChild>
                <Button
                  size="sm"
                  variant="destructive"
                  disabled={archiving}
                  className="inline-flex items-center gap-1"
                >
                  <Trash2Icon className="h-4 w-4" />
                  Supprimer
                </Button>
              </AlertDialogTrigger>
              <AlertDialogContent>
                <AlertDialogHeader>
                  <AlertDialogTitle>Supprimer ce client ?</AlertDialogTitle>
                  <AlertDialogDescription>
                    Cette action est irréversible. Les adresses associées
                    resteront dans la base.
                  </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter>
                  <AlertDialogCancel>Annuler</AlertDialogCancel>
                  <AlertDialogAction
                    variant="destructive"
                    onClick={handleArchive}
                  >
                    Supprimer
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
            <CardTitle>Informations</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1 text-sm">
            <p>
              <span className="font-medium">Prénom:</span> {client.first_name}
            </p>
            <p>
              <span className="font-medium">Nom:</span> {client.last_name}
            </p>
            <p>
              <span className="font-medium">Email:</span> {client.email}
            </p>
            {client.phone && (
              <p>
                <span className="font-medium">Téléphone:</span> {client.phone}
              </p>
            )}
            {client.company && (
              <p>
                <span className="font-medium">Société:</span> {client.company}
              </p>
            )}
            {client.siren && (
              <p>
                <span className="font-medium">SIREN:</span> {client.siren}
              </p>
            )}
            {client.vat && (
              <p>
                <span className="font-medium">TVA:</span> {client.vat}
              </p>
            )}
          </CardContent>
        </Card>

        <div>
          <h3 className="mb-2 text-sm font-medium">Adresses</h3>
          <AddressesTable ownerType="client" ownerId={uuid} />
        </div>
      </CardContent>
    </Card>
  );
}
