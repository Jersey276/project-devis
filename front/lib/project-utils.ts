import type { BackendClient, ProjectStatus } from "@/types/backend";

export const PROJECT_STATUS_ITEMS: { value: string; label: string }[] = [
  { value: "active", label: "Actif" },
  { value: "completed", label: "Terminé" },
  { value: "archived", label: "Archivé" },
];

export function clientLabel(c: BackendClient): string {
  const name = `${c.first_name} ${c.last_name}`.trim();
  return name || c.company || c.client_id;
}

export function clientName(clientId: string, clients: BackendClient[], fallback = "—"): string {
  if (!clientId) return fallback;
  const c = clients.find((cl) => cl.client_id === clientId);
  if (!c) return fallback;
  return clientLabel(c);
}
