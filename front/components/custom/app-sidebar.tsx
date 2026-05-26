"use client";

import { useMemo } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import {
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

type NavKey = "quote" | "invoices" | "clients" | "countries" | "taxes" | "templates" | "test";

type SidebarItem = {
  key: NavKey;
  url: string;
  icon: LucideIcon;
  // Modes in which this entry is visible. Omit to show in every mode.
  modes?: UserMode[];
  // Marker for entries that will be gated by the upcoming roles/permissions system.
  temp?: boolean;
};

const items: SidebarItem[] = [
  {
    key: "quote",
    url: "/quote",
    icon: QuoteIcon,
  },
  // {
  //   key: "invoices",
  //   url: "/invoice",
  //   icon: ReceiptEuroIcon,
  //   modes: ["provider"],
  // },
  {
    key: "clients",
    url: "/clients",
    icon: QuoteIcon,
    modes: ["provider"],
  },
  {
    key: "countries",
    url: "/countries",
    icon: GlobeIcon,
    modes: ["provider"],
    temp: true,
  },
  {
    key: "taxes",
    url: "/taxes",
    icon: PercentIcon,
    modes: ["provider"],
    temp: true,
  },
  {
    key: "templates",
    url: "/templates",
    icon: WrenchIcon,
    modes: ["provider"],
  },
  // {
  //   key: "test",
  //   url: "/test",
  //   icon: WrenchIcon,
  //   modes: ["provider"],
  // },
];

export default function AppSidebar() {
  const { mode } = useMode();
  const t = useTranslations("nav");
  const visibleItems = useMemo(
    () => items.filter((item) => !item.modes || item.modes.includes(mode)),
    [mode],
  );
  return (
    <Sidebar data-mode={mode}>
      <SidebarContent className="bg-primary-foreground text-primary">
        <SidebarGroup>
          <SidebarGroupLabel>{t("appGroupLabel")}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {visibleItems.map((item) => (
                <SidebarMenuItem key={item.key}>
                  <SidebarMenuButton asChild>
                    <Link href={item.url}>
                      <item.icon />
                      <span>{t(item.key)}</span>
                    </Link>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
      <SidebarFooter className="bg-primary-foreground text-primary">
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
