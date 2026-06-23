import { apiFetch, type ApiResult } from "@/lib/api";

export type CreateProjectPayload = {
  name: string;
  clientId?: string;
};

export type UpdateProjectPayload = {
  name: string;
  clientId?: string;
  status?: string;
};

export async function listProjects(queryString?: string, signal?: AbortSignal): Promise<ApiResult> {
  const path = queryString ? `/api/projects?${queryString}` : "/api/projects";
  return apiFetch(path, { signal });
}

export async function createProject(payload: CreateProjectPayload): Promise<ApiResult> {
  return apiFetch("/api/projects", {
    method: "POST",
    body: JSON.stringify({ name: payload.name, client_id: payload.clientId ?? "" }),
  });
}

export async function getProject(projectId: string): Promise<ApiResult> {
  return apiFetch(`/api/projects/${encodeURIComponent(projectId)}`);
}

export async function getProjectDetail(projectId: string, signal?: AbortSignal): Promise<ApiResult> {
  return apiFetch(`/api/projects/${encodeURIComponent(projectId)}/detail`, { signal });
}

export async function updateProject(projectId: string, payload: UpdateProjectPayload): Promise<ApiResult> {
  return apiFetch(`/api/projects/${encodeURIComponent(projectId)}`, {
    method: "PUT",
    body: JSON.stringify({
      name: payload.name,
      client_id: payload.clientId ?? "",
      status: payload.status ?? "active",
    }),
  });
}

export async function deleteProject(projectId: string): Promise<ApiResult> {
  return apiFetch(`/api/projects/${encodeURIComponent(projectId)}`, { method: "DELETE" });
}

export async function addQuoteToProject(projectId: string, quoteId: string): Promise<ApiResult> {
  return apiFetch(`/api/projects/${encodeURIComponent(projectId)}/quotes`, {
    method: "POST",
    body: JSON.stringify({ quote_id: quoteId }),
  });
}

export async function removeQuoteFromProject(projectId: string, quoteId: string): Promise<ApiResult> {
  return apiFetch(
    `/api/projects/${encodeURIComponent(projectId)}/quotes/${encodeURIComponent(quoteId)}`,
    { method: "DELETE" },
  );
}
