"use client";

import Link from "next/link";
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

type SidebarItem = {
  title: string;
  url: string;
  icon: LucideIcon;
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
  },
  {
    title: "Clients",
    url: "/clients",
    icon: QuoteIcon,
  },
  {
    title: "Pays",
    url: "/countries",
    icon: GlobeIcon,
    temp: true,
  },
  {
    title: "Taxes",
    url: "/taxes",
    icon: PercentIcon,
    temp: true,
  },
  {
    title: "Test",
    url: "/test",
    icon: WrenchIcon,
  },
];

export default function AppSidebar() {
  return (
    <Sidebar>
      <SidebarContent className="bg-primary-foreground text-primary">
        <SidebarGroup>
          <SidebarGroupLabel>Application</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item) => (
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
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
