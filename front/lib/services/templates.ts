import { apiFetch, type ApiResult } from "@/lib/api";
import type {
  BackendTemplate,
  BackendTemplateLine,
  BackendTemplateType,
  BackendTemplateTargetResource,
} from "@/types/backend";

export type CreateTemplatePayload = {
  templateType: BackendTemplateType;
  targetResource: BackendTemplateTargetResource;
  name: string;
};

export async function listTemplates(
  options: { archived?: boolean; type?: BackendTemplateType } = {},
): Promise<ApiResult> {
  const params = new URLSearchParams();
  if (options.archived) params.set("archived", "true");
  if (options.type) params.set("type", options.type);
  const qs = params.toString();
  return apiFetch(`/api/templates${qs ? `?${qs}` : ""}`);
}

export async function createTemplate(
  payload: CreateTemplatePayload,
): Promise<ApiResult> {
  return apiFetch("/api/templates", {
    method: "POST",
    body: JSON.stringify({
      template_type: payload.templateType,
      target_resource: payload.targetResource,
      name: payload.name,
    }),
  });
}

export async function getTemplate(templateId: string): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}`);
}

export type UpdateTemplatePayload = {
  name?: string;
  targetResource?: BackendTemplateTargetResource;
  payload?: Record<string, unknown>;
  payloadVersion?: number;
};

export async function updateTemplate(
  templateId: string,
  payload: UpdateTemplatePayload,
): Promise<ApiResult> {
  const body: Record<string, unknown> = {};
  if (payload.name !== undefined) body.name = payload.name;
  if (payload.targetResource !== undefined)
    body.target_resource = payload.targetResource;
  if (payload.payload !== undefined) body.payload = payload.payload;
  if (payload.payloadVersion !== undefined)
    body.payload_version = payload.payloadVersion;
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}`, {
    method: "PUT",
    body: JSON.stringify(body),
  });
}

export async function deleteTemplate(templateId: string): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}`, {
    method: "DELETE",
  });
}

export async function archiveTemplate(templateId: string): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}/archive`, {
    method: "POST",
  });
}

export async function restoreTemplate(templateId: string): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}/restore`, {
    method: "POST",
  });
}

// ─── Template lines ──────────────────────────────────────────────────────────

export type TemplateLineDraft = {
  type: string;
  name: string;
  quantity: number;
  unit?: string;
  unitPriceEuros: number;
  position: number;
  taxId: number | null;
};

function toCents(euros: number): number {
  return Math.round(euros * 100);
}

function toLinePayload(draft: TemplateLineDraft) {
  return {
    type: draft.type,
    name: draft.name,
    quantity: String(draft.quantity),
    unit: draft.unit ?? "",
    unit_price: toCents(draft.unitPriceEuros),
    data: {},
    position: draft.position,
    tax_id: draft.taxId ?? 0,
  };
}

export async function listTemplateLines(
  templateId: string,
): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}/lines`);
}

export async function createTemplateLine(
  templateId: string,
  draft: TemplateLineDraft,
): Promise<ApiResult> {
  return apiFetch(`/api/templates/${encodeURIComponent(templateId)}/lines`, {
    method: "POST",
    body: JSON.stringify(toLinePayload(draft)),
  });
}

export async function updateTemplateLine(
  templateId: string,
  lineId: string,
  draft: TemplateLineDraft,
): Promise<ApiResult> {
  return apiFetch(
    `/api/templates/${encodeURIComponent(templateId)}/lines/${encodeURIComponent(lineId)}`,
    {
      method: "PUT",
      body: JSON.stringify(toLinePayload(draft)),
    },
  );
}

export async function deleteTemplateLine(
  templateId: string,
  lineId: string,
): Promise<ApiResult> {
  return apiFetch(
    `/api/templates/${encodeURIComponent(templateId)}/lines/${encodeURIComponent(lineId)}`,
    { method: "DELETE" },
  );
}

export type { BackendTemplate, BackendTemplateLine };
