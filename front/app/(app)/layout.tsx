import { cookies } from "next/headers";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/custom/app-sidebar";
import { ModeProvider, type UserMode } from "@/lib/mode-context";
import { AuthProvider } from "@/lib/auth-context";
import { AUTH_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/auth-constants";
import { redirect } from "next/navigation";
import type { AuthContext } from "@/lib/access";
import ConsentGate from "@/components/consent/consent-gate";
import { CONSENT_VERSIONS, type ConsentType } from "@/lib/consent-versions";

const gatewayUrl =
  process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "http://devis-gateway:8080";

// Attempts a server-side token refresh. On success, propagates the new cookies
// to the browser via Next.js cookies() API and returns the new access token.
async function tryRefreshSSR(cookieHeader: string): Promise<string | null> {
  try {
    const res = await fetch(`${gatewayUrl}/api/auth/refresh`, {
      method: "POST",
      headers: { Cookie: cookieHeader, Accept: "application/json" },
      cache: "no-store",
    });
    if (!res.ok) return null;
    const cookieStore = await cookies();
    for (const raw of res.headers.getSetCookie()) {
      // Parse name=value and optional attributes (Path, HttpOnly, Max-Age, SameSite…).
      const parts = raw.split(/;\s*/);
      const [nameVal, ...attrs] = parts;
      const eqIdx = nameVal.indexOf("=");
      if (eqIdx === -1) continue;
      const name = nameVal.slice(0, eqIdx);
      const value = nameVal.slice(eqIdx + 1);
      const attrMap: Record<string, string | boolean> = {};
      for (const attr of attrs) {
        const [k, v] = attr.split("=");
        attrMap[k.toLowerCase()] = v ?? true;
      }
      cookieStore.set(name, value, {
        path: (attrMap["path"] as string) ?? "/",
        httpOnly: "httponly" in attrMap,
        secure: "secure" in attrMap,
        maxAge: attrMap["max-age"] ? Number(attrMap["max-age"]) : undefined,
        sameSite: (attrMap["samesite"] as "lax" | "strict" | "none") ?? "lax",
      });
      if (name === AUTH_TOKEN_COOKIE) return value;
    }
    return null;
  } catch {
    return null;
  }
}

export default async function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const cookieHeader = cookieStore
    .getAll()
    .map((c) => `${c.name}=${c.value}`)
    .join("; ");

  // If the access token cookie is missing but a refresh token exists, attempt a
  // silent SSR refresh before redirecting to login. This avoids a spurious login
  // redirect when the 2-minute access token expires between page navigations.
  if (!cookieStore.get(AUTH_TOKEN_COOKIE)) {
    if (!cookieStore.get(REFRESH_TOKEN_COOKIE)) {
      redirect("/login");
    }
    const newToken = await tryRefreshSSR(cookieHeader);
    if (!newToken) {
      redirect("/login");
    }
    // Cookie was set above; rebuild the header with the new access token so the
    // /me call below uses the fresh token.
    cookieStore.set(AUTH_TOKEN_COOKIE, newToken);
  }

  const freshCookieHeader = cookieStore
    .getAll()
    .map((c) => `${c.name}=${c.value}`)
    .join("; ");

  let serverAuth: AuthContext | null = null;
  let authOk = false;
  let outdatedConsents: ConsentType[] = [];

  const fetchOptions = { headers: { Cookie: freshCookieHeader }, cache: "no-store" } as const;

  const [meRes, consentRes] = await Promise.all([
    fetch(`${gatewayUrl}/api/auth/me`, fetchOptions).catch(() => null),
    fetch(`${gatewayUrl}/api/consent/status`, fetchOptions).catch(() => null),
  ]);

  if (meRes?.ok) {
    try {
      const data = await meRes.json();
      if (data.auth?.email_verified === false) {
        redirect("/verify-email");
      }
      if (data.success === true) {
        serverAuth = (data.auth ?? null) as AuthContext | null;
        authOk = true;
      }
    } catch {
      // Gateway indisponible : fail open pour ne pas bloquer l'accès.
    }
  }

  if (authOk && consentRes?.ok) {
    try {
      const consentData = await consentRes.json();
      const accepted: Record<string, string> = {};
      for (const entry of consentData.consents ?? []) {
        accepted[entry.type] = entry.version;
      }
      outdatedConsents = (Object.keys(CONSENT_VERSIONS) as ConsentType[]).filter(
        (type) => accepted[type] !== CONSENT_VERSIONS[type]
      );
    } catch {
      // Fail open — ne bloque pas l'accès si le service est indisponible.
    }
  }

  const rawMode = cookieStore.get("user-mode")?.value;
  const initialMode: UserMode = rawMode === "customer" ? "customer" : "provider";
  return (
    <ModeProvider initialMode={initialMode}>
      <AuthProvider auth={serverAuth} ok={authOk}>
        <SidebarProvider>
          <AppSidebar />
          <SidebarInset>
            <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
              <SidebarTrigger className="-ml-1" />
            </header>
            <main className="p-4">{children}</main>
          </SidebarInset>
        </SidebarProvider>
        {outdatedConsents.length > 0 && <ConsentGate outdated={outdatedConsents} />}
      </AuthProvider>
    </ModeProvider>
  );
}
