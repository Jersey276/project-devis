// Must stay in sync with backend field validation codes (see backend/auth/actions/errors.go).
export const FIELD_VALIDATION_MESSAGES: Record<number, string> = {
  1: "Ce champ est requis.",
  2: "Format invalide.",
  3: "Trop court (12 caractères minimum).",
  4: "Cette adresse email est déjà utilisée.",
};

export type FieldErrors = Record<string, string[]>;

export type ApiFieldError = { field: string; error_code: number[] };

// Some services (users gateway) return already-localized per-field messages under
// `errors` instead of the auth-style coded `field_errors`. Both are supported.
export type ApiFieldMessage = { field: string; message: string };

export type ApiBody = {
  success: boolean;
  message?: string;
  code?: number;
  field_errors?: ApiFieldError[];
  errors?: ApiFieldMessage[];
  [key: string]: unknown;
};

export type ApiResult = {
  ok: boolean;
  status: number;
  body: ApiBody;
};

const CODE_SESSION_INVALIDATED = 1008;

// Endpoints that must not trigger the refresh-and-retry loop (would recurse or mask login failures).
// /api/auth/password/update returns 401 when the current password is wrong — not a session
// expiry — so refreshing the token and retrying would just log the user out on a typo.
const REFRESH_SKIP_PATHS = new Set([
  "/api/auth/refresh",
  "/api/auth/login",
  "/api/auth/logout",
  "/api/auth/password/update",
  "/api/auth/password/reset",
  "/api/auth/password/confirm-reset",
]);

// Coalesces concurrent 401s onto a single /api/auth/refresh call.
let refreshPromise: Promise<boolean> | null = null;
let sessionInvalidationPromise: Promise<void> | null = null;

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

function redirectToLogin() {
  if (typeof window !== "undefined") {
    const next = `${window.location.pathname}${window.location.search}${window.location.hash}`;
    window.location.href = `/login?next=${encodeURIComponent(next)}`;
  }
}

async function readResponseCode(
  response: Response,
): Promise<number | undefined> {
  try {
    const body = (await response.clone().json()) as ApiBody;
    return body.code;
  } catch {
    return undefined;
  }
}

async function handleSessionInvalidation(): Promise<void> {
  if (!sessionInvalidationPromise) {
    sessionInvalidationPromise = (async () => {
      try {
        await fetch("/api/auth/logout", {
          method: "POST",
          credentials: "include",
          headers: { Accept: "application/json" },
        });
      } catch {
        // Best effort only: even if logout fails, we still redirect.
      } finally {
        redirectToLogin();
        sessionInvalidationPromise = null;
      }
    })();
  }
  await sessionInvalidationPromise;
}

// Runs `doFetch`, retries once after a successful /api/auth/refresh on a 401,
// and triggers a redirect to /login when refresh fails. Returns null when the
// caller should bail (redirect was triggered). When `skipRefresh` is true, the
// 401 is returned verbatim — used by auth endpoints that handle 401 themselves.
export async function fetchWithRefresh(
  doFetch: () => Promise<Response>,
  opts: { skipRefresh?: boolean } = {},
): Promise<Response | null> {
  let res = await doFetch();
  if (res.status !== 401) return res;

  const code = await readResponseCode(res);
  if (code === CODE_SESSION_INVALIDATED) {
    await handleSessionInvalidation();
    return null;
  }

  if (opts.skipRefresh) return res;

  const refreshed = await attemptRefresh();
  if (!refreshed) {
    redirectToLogin();
    return null;
  }
  res = await doFetch();
  if (res.status === 401) {
    redirectToLogin();
    return null;
  }
  return res;
}

export function readUserModeCookie(): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.split("; ").find((c) => c.startsWith("user-mode="));
  return match?.split("=")[1];
}

export async function apiFetch(
  path: string,
  init?: RequestInit,
): Promise<ApiResult> {
  const clientMode = readUserModeCookie() === "customer";
  const doFetch = () =>
    fetch(path, {
      ...init,
      credentials: "include",
      headers: {
        Accept: "application/json",
        ...(init?.body ? { "Content-Type": "application/json" } : {}),
        ...(clientMode ? { "X-Client-Mode": "customer" } : {}),
        ...(init?.headers ?? {}),
      },
    });

  const response = await fetchWithRefresh(doFetch, {
    skipRefresh: REFRESH_SKIP_PATHS.has(path),
  });
  if (!response) return { ok: false, status: 401, body: { success: false } };

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
  // Auth-style coded errors, translated client-side.
  if (Array.isArray(body.field_errors)) {
    for (const entry of body.field_errors) {
      errors[entry.field] = entry.error_code.map(
        (code) =>
          FIELD_VALIDATION_MESSAGES[code] ?? `Erreur de validation (${code}).`,
      );
    }
  }
  // Users-style errors carry an already-localized message per field.
  if (Array.isArray(body.errors)) {
    for (const entry of body.errors) {
      (errors[entry.field] ??= []).push(entry.message);
    }
  }
  return errors;
}

export function toErrorProps(messages: string[] | undefined) {
  return messages?.map((message) => ({ message }));
}
