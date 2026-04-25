// Must stay in sync with backend field validation codes (see backend/auth/actions/errors.go).
export const FIELD_VALIDATION_MESSAGES: Record<number, string> = {
  1: "Ce champ est requis.",
  2: "Format invalide.",
  3: "Trop court (8 caractères minimum).",
  4: "Cette adresse email est déjà utilisée.",
};

export type FieldErrors = Record<string, string[]>;

export type ApiFieldError = { field: string; error_code: number[] };

export type ApiBody = {
  success: boolean;
  message?: string;
  code?: number;
  field_errors?: ApiFieldError[];
  [key: string]: unknown;
};

export type ApiResult = {
  ok: boolean;
  status: number;
  body: ApiBody;
};

// Endpoints that must not trigger the refresh-and-retry loop (would recurse or mask login failures).
const REFRESH_SKIP_PATHS = new Set([
  "/api/auth/refresh",
  "/api/auth/login",
  "/api/auth/logout",
]);

// Coalesces concurrent 401s onto a single /api/auth/refresh call.
let refreshPromise: Promise<boolean> | null = null;

function attemptRefresh(): Promise<boolean> {
  if (refreshPromise) return refreshPromise;
  refreshPromise = (async () => {
    try {
      const response = await fetch("/api/auth/refresh", {
        method: "POST",
        credentials: "include",
        headers: { Accept: "application/json" },
      });
      return response.ok;
    } catch {
      return false;
    } finally {
      refreshPromise = null;
    }
  })();
  return refreshPromise;
}

function redirectToLogin(): ApiResult {
  if (typeof window !== "undefined") {
    window.location.href = "/login";
  }
  return { ok: false, status: 401, body: { success: false } };
}

export async function apiFetch(
  path: string,
  init?: RequestInit,
): Promise<ApiResult> {
  const doFetch = () =>
    fetch(path, {
      ...init,
      credentials: "include",
      headers: {
        Accept: "application/json",
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...(init?.headers ?? {}),
      },
    });

  let response = await doFetch();

  if (response.status === 401 && !REFRESH_SKIP_PATHS.has(path)) {
    const refreshed = await attemptRefresh();
    if (!refreshed) return redirectToLogin();
    response = await doFetch();
    if (response.status === 401) return redirectToLogin();
  }

  let body: ApiBody;
  try {
    body = (await response.json()) as ApiBody;
  } catch {
    body = { success: false };
  }
  return { ok: response.ok, status: response.status, body };
}

export function fieldErrorsFromBody(body: ApiBody): FieldErrors {
  const errors: FieldErrors = {};
  if (!Array.isArray(body.field_errors)) return errors;
  for (const entry of body.field_errors) {
    errors[entry.field] = entry.error_code.map(
      (code) =>
        FIELD_VALIDATION_MESSAGES[code] ?? `Erreur de validation (${code}).`,
    );
  }
  return errors;
}

export function toErrorProps(messages: string[] | undefined) {
  return messages?.map((message) => ({ message }));
}
