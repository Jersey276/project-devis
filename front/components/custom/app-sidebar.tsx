"use client";

import { useMemo } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  EyeIcon,
  GlobeIcon,
  PercentIcon,
  QuoteIcon,
  ReceiptEuroIcon,
  WrenchIcon,
  type LucideIcon,
} from "lucide-react";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "../ui/sidebar";
import UserMenu from "../user/user-menu";
import { useMode, type UserMode } from "@/lib/mode-context";

type SidebarItem = {
  title: string;
  url: string;
  icon: LucideIcon;
  // Modes in which this entry is visible. Omit to show in every mode.
  modes?: UserMode[];
  // Marker for entries that will be gated by the upcoming roles/permissions system.
  temp?: boolean;
};

const items: SidebarItem[] = [
  {
    title: "Devis",
    url: "/quote",
    icon: QuoteIcon,
  },
  {
    title: "Factures",
    url: "/invoice",
    icon: ReceiptEuroIcon,
    modes: ["provider"],
  },
  {
    title: "Clients",
    url: "/clients",
    icon: QuoteIcon,
    modes: ["provider"],
  },
  {
    title: "Pays",
    url: "/countries",
    icon: GlobeIcon,
    modes: ["provider"],
    temp: true,
  },
  {
    title: "Taxes",
    url: "/taxes",
    icon: PercentIcon,
    modes: ["provider"],
    temp: true,
  },
  {
    title: "Test",
    url: "/test",
    icon: WrenchIcon,
    modes: ["provider"],
  },
];

export default function AppSidebar() {
  const router = useRouter();
  const pathname = usePathname();
  const { mode, setMode, isCustomer } = useMode();
  const visibleItems = useMemo(
    () => items.filter((item) => !item.modes || item.modes.includes(mode)),
    [mode],
  );

  // When entering customer mode, redirect to /quote only if the current
  // route is a provider-only sidebar path (e.g. /clients). Switching back to
  // provider never redirects — every route is visible. Avoiding a no-op
  // router.replace here matters: replacing the URL with itself re-runs the
  // (app) layout and briefly tears the sidebar down.
  function handleModeToggle() {
    const next: UserMode = isCustomer ? "provider" : "customer";
    setMode(next);
    if (next !== "customer") return;
    const onProviderOnlyPath = items.some(
      (item) =>
        item.modes?.length === 1 &&
        item.modes[0] === "provider" &&
        (pathname === item.url || pathname.startsWith(`${item.url}/`)),
    );
    if (onProviderOnlyPath) router.replace("/quote");
  }
  return (
    <Sidebar data-mode={mode}>
      <SidebarContent className="bg-primary-foreground text-primary">
        <SidebarGroup>
          <SidebarGroupLabel>Application</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {visibleItems.map((item) => (
                <SidebarMenuItem key={item.title}>
                  <SidebarMenuButton asChild>
                    <Link href={item.url}>
                      <item.icon />
                      <span>{item.title}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="bg-primary-foreground text-primary">
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              data-slot="mode-toggle"
              data-active={isCustomer ? "true" : undefined}
              aria-pressed={isCustomer}
              onClick={handleModeToggle}
            >
              <EyeIcon />
              <span>Mode client</span>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
