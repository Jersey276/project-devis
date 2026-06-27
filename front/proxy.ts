import { NextRequest, NextResponse } from "next/server";
import { AUTH_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/auth-constants";
import { NEXT_PARAM } from "@/lib/auth-utils";

const PUBLIC_PATHS = new Set([
  "/login",
  "/register",
  "/forget-password",
  "/reset-password",
  "/accept-invitation",
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

  // Access token absent but refresh token present: let the SSR layout handle
  // the refresh via tryRefreshSSR(). Doing it here too would consume the
  // refresh token before the layout reads it, causing a double-refresh race
  // that ends in a spurious redirect to /login.
  return NextResponse.next();
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
