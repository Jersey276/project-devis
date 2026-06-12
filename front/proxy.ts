import { NextRequest, NextResponse } from "next/server";
import { AUTH_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/auth-constants";
import { NEXT_PARAM } from "@/lib/auth-utils";

const REFRESH_TIMEOUT_MS = 5000;

const PUBLIC_PATHS = new Set([
  "/login",
  "/register",
  "/forget-password",
  "/reset-password",
]);

export async function proxy(request: NextRequest) {
  if (PUBLIC_PATHS.has(request.nextUrl.pathname)) {
    return NextResponse.next();
  }

  const cookies = request.cookies;

  if (cookies.get(AUTH_TOKEN_COOKIE)) {
    return NextResponse.next();
  }

  if (!cookies.get(REFRESH_TOKEN_COOKIE)) {
    return redirectToLogin(request);
  }

  const refreshResponse = await fetch(
    new URL("/api/auth/refresh", request.url),
    {
      method: "POST",
      headers: {
        cookie: request.headers.get("cookie") ?? "",
        accept: "application/json",
      },
      signal: AbortSignal.timeout(REFRESH_TIMEOUT_MS),
    },
  ).catch(() => null);

  if (!refreshResponse?.ok) {
    return redirectToLogin(request);
  }

  const refreshed = (await refreshResponse.json().catch(() => null)) as {
    token?: string;
    refresh_token?: string;
  } | null;
  if (refreshed?.token && refreshed?.refresh_token) {
    request.cookies.set(AUTH_TOKEN_COOKIE, refreshed.token);
    request.cookies.set(REFRESH_TOKEN_COOKIE, refreshed.refresh_token);
  }

  const response = NextResponse.next({ request });
  for (const setCookie of refreshResponse.headers.getSetCookie()) {
    response.headers.append("set-cookie", setCookie);
  }
  return response;
}

function redirectToLogin(request: NextRequest) {
  const loginUrl = new URL("/login", request.url);
  loginUrl.searchParams.set(
    NEXT_PARAM,
    request.nextUrl.pathname + request.nextUrl.search,
  );
  return NextResponse.redirect(loginUrl);
}

export const config = {
  // Protect every route except the auth pages, API (would loop the refresh call),
  // Next.js internals, and static assets.
  matcher: [
    "/((?!login|register|forget-password|reset-password|api/|_next/|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp|ico|css|js|map)$).*)",
  ],
};
