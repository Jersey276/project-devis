import { cookies } from "next/headers";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import AppSidebar from "@/components/custom/app-sidebar";
import { ModeProvider, type UserMode } from "@/lib/mode-context";

export default async function AppLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const cookieStore = await cookies();
  const modeCookie = cookieStore.get("app.user-mode")?.value;
  const initialMode: UserMode =
    modeCookie === "customer" || modeCookie === "provider"
      ? modeCookie
      : "provider";
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
