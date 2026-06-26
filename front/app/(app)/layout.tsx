import { cookies } from "next/headers";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/custom/app-sidebar";
import { ModeProvider, type UserMode } from "@/lib/mode-context";
import { AuthProvider } from "@/lib/auth-context";
import { AUTH_TOKEN_COOKIE } from "@/lib/auth-constants";
import { redirect } from "next/navigation";
import type { AuthContext } from "@/lib/access";

const gatewayUrl =
  process.env.NODE_ENV === "development"
    ? "http://localhost:8080"
    : "http://devis-gateway:8080";

export default async function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  if (!cookieStore.get(AUTH_TOKEN_COOKIE)) {
    redirect("/login");
  }

  let serverAuth: AuthContext | null = null;
  let authOk = false;

  try {
    const cookieHeader = cookieStore
      .getAll()
      .map((c) => `${c.name}=${c.value}`)
      .join("; ");
    const meRes = await fetch(`${gatewayUrl}/api/auth/me`, {
      headers: { Cookie: cookieHeader },
      cache: "no-store",
    });
    if (meRes.ok) {
      const data = await meRes.json();
      if (data.auth?.email_verified === false) {
        redirect("/verify-email");
      }
      if (data.success === true) {
        serverAuth = (data.auth ?? null) as AuthContext | null;
        authOk = true;
      }
    }
  } catch {
    // Gateway indisponible : fail open pour ne pas bloquer l'accès.
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
      </AuthProvider>
    </ModeProvider>
  );
}
