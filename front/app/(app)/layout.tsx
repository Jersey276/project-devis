import { cookies } from "next/headers";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/custom/app-sidebar";
import { ModeProvider, type UserMode } from "@/lib/mode-context";
import { AUTH_TOKEN_COOKIE } from "@/lib/auth-constants";
import { redirect } from "next/navigation";

export default async function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  if (!cookieStore.get(AUTH_TOKEN_COOKIE)) {
    redirect("/login");
  }
  const rawMode = cookieStore.get("user-mode")?.value;
  const initialMode: UserMode = rawMode === "customer" ? "customer" : "provider";
  return (
    <ModeProvider initialMode={initialMode}>
      <SidebarProvider>
        <AppSidebar />
        <SidebarInset>
          <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
            <SidebarTrigger className="-ml-1" />
          </header>
          <main className="p-4">{children}</main>
        </SidebarInset>
      </SidebarProvider>
    </ModeProvider>
  );
}
