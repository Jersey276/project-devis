import { apiFetch, type ApiResult } from "@/lib/api";

function commentsUrl(quoteId: string, lineId: string, commentId?: string): string {
  const base = `/api/quotes/${encodeURIComponent(quoteId)}/lines/${encodeURIComponent(lineId)}/comments`;
  return commentId ? `${base}/${encodeURIComponent(commentId)}` : base;
}

export async function listComments(quoteId: string, lineId: string): Promise<ApiResult> {
  return apiFetch(commentsUrl(quoteId, lineId));
}

export async function createComment(quoteId: string, lineId: string, body: string, authorName?: string): Promise<ApiResult> {
  return apiFetch(commentsUrl(quoteId, lineId), {
    method: "POST",
    body: JSON.stringify({ body, author_name: authorName ?? "" }),
  });
}

export async function updateComment(quoteId: string, lineId: string, commentId: string, body: string): Promise<ApiResult> {
  return apiFetch(commentsUrl(quoteId, lineId, commentId), {
    method: "PUT",
    body: JSON.stringify({ body }),
  });
}

export async function deleteComment(quoteId: string, lineId: string, commentId: string): Promise<ApiResult> {
  return apiFetch(commentsUrl(quoteId, lineId, commentId), { method: "DELETE" });
}
