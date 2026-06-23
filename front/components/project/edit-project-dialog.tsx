"use client";

import { useEffect, useState } from "react";
import {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogContent,
  ResponsiveDialogFooter,
  ResponsiveDialogHeader,
  ResponsiveDialogTitle,
} from "@/components/custom/responsive-dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { SelectCombobox } from "@/components/ui/select-combobox";
import { toast } from "sonner";
import { updateProject } from "@/lib/services/projects";
import { listClients } from "@/lib/services/clients";
import { clientLabel, PROJECT_STATUS_ITEMS } from "@/lib/project-utils";
import type { BackendClient, ProjectStatus } from "@/types/backend";

export type EditableProject = {
  project_id: string;
  name: string;
  client_id: string;
  status: ProjectStatus;
};

type Props = {
  project: EditableProject;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onUpdated?: () => void;
};

export default function EditProjectDialog({ project, open, onOpenChange, onUpdated }: Props) {
  const [name, setName] = useState(project.name);
  const [clientId, setClientId] = useState(project.client_id ?? "");
  const [status, setStatus] = useState(project.status);
  const [clients, setClients] = useState<BackendClient[]>([]);
  const [nameError, setNameError] = useState("");
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    if (!open) return;
    setName(project.name);
    setClientId(project.client_id ?? "");
    setStatus(project.status);
    listClients().then(({ ok, body }) => {
      if (ok && Array.isArray(body.clients)) setClients(body.clients);
    });
  }, [open, project]);

  async function handleSubmit() {
    if (!name.trim()) { setNameError("Le nom est requis."); return; }
    setNameError("");
    setBusy(true);
    const { ok, body } = await updateProject(project.project_id, {
      name: name.trim(),
      clientId: clientId || undefined,
      status,
    });
    setBusy(false);
    if (!ok) { toast.error(body?.message ?? "La mise à jour a échoué."); return; }
    toast.success("Projet mis à jour.");
    onOpenChange(false);
    onUpdated?.();
  }

  return (
    <ResponsiveDialog open={open} onOpenChange={onOpenChange}>
      <ResponsiveDialogContent>
        <ResponsiveDialogHeader>
          <ResponsiveDialogTitle>Modifier le projet</ResponsiveDialogTitle>
        </ResponsiveDialogHeader>
        <ResponsiveDialogBody>
          <div className="flex flex-col gap-4">
          <Field>
            <FieldLabel>Nom du projet</FieldLabel>
            <Input value={name} onChange={(e) => setName(e.target.value)} />
            {nameError && <FieldError>{nameError}</FieldError>}
          </Field>
          <Field>
            <FieldLabel>Client</FieldLabel>
            <Combobox value={clientId} onValueChange={setClientId}>
              <ComboboxInput placeholder="Rechercher un client…" />
              <ComboboxContent>
                <ComboboxList>
                  <ComboboxEmpty>Aucun client.</ComboboxEmpty>
                  {clients.map((c) => (
                    <ComboboxItem key={c.client_id} value={c.client_id}>
                      {clientLabel(c)}
                    </ComboboxItem>
                  ))}
                </ComboboxList>
              </ComboboxContent>
            </Combobox>
          </Field>
          <Field>
            <FieldLabel>Statut</FieldLabel>
            <SelectCombobox
              items={PROJECT_STATUS_ITEMS}
              value={status}
              onValueChange={(v) => setStatus(v as typeof status)}
              placeholder="Statut"
            />
          </Field>
          </div>
        </ResponsiveDialogBody>
        <ResponsiveDialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>
            Annuler
          </Button>
          <Button onClick={handleSubmit} disabled={busy}>
            Enregistrer
          </Button>
        </ResponsiveDialogFooter>
      </ResponsiveDialogContent>
    </ResponsiveDialog>
  );
}
