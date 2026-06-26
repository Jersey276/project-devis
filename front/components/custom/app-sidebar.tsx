"use client";

import { useMemo, useState } from "react";
import Link from "next/link";
import { useTranslations } from "next-intl";
import {
  ShieldIcon,
  UserIcon,
  GlobeIcon,
  PercentIcon,
  QuoteIcon,
  UsersIcon,
  WrenchIcon,
  CreditCardIcon,
  CoinsIcon,
  ReceiptEuroIcon,
  FolderIcon,
  BuildingIcon,
  LayoutDashboardIcon,
  MailIcon,
  type LucideIcon,
} from "lucide-react";
import { Button } from "../ui/button";
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
import { useAuth } from "@/lib/auth-context";
import { canUsePaidFeatures, isSuperAdmin } from "@/lib/access";
import { cn } from "@/lib/utils";

type NavKey =
  | "dashboard"
  | "adminDashboard"
  | "project"
  | "quote"
  | "schedule"
  | "invoices"
  | "creditNotes"
  | "clients"
  | "clientProfile"
  | "fees"
  | "users"
  | "countries"
  | "taxes"
  | "templates"
  | "subscriptions"
  | "logs"
  | "emailLogs"
  | "test";

type SidebarItem = {
  key: NavKey;
  url: string;
  icon: LucideIcon;
  modes?: UserMode[];
  temp?: boolean;
  adminOnly?: boolean;
  premium?: boolean;
};

type SidebarView = "user" | "admin";

const items: SidebarItem[] = [
  {
    key: "dashboard",
    url: "/",
    icon: LayoutDashboardIcon,
  },
  {
    key: "project",
    url: "/projects",
    icon: FolderIcon,
  },
  {
    key: "quote",
    url: "/quote",
    icon: QuoteIcon,
  },
  {
    key: "schedule",
    url: "/schedule",
    icon: QuoteIcon,
    modes: ["provider"],
    premium: true,
  },
  {
    key: "invoices",
    url: "/invoice",
    icon: ReceiptEuroIcon,
    modes: ["provider"],
    premium: true,
  },
  {
    key: "creditNotes",
    url: "/credit-note",
    icon: ReceiptEuroIcon,
    modes: ["provider"],
    premium: true,
  },
  {
    key: "clients",
    url: "/clients",
    icon: QuoteIcon,
    modes: ["provider"],
  },
  {
    key: "clientProfile",
    url: "/client-profile",
    icon: UserIcon,
    modes: ["customer"],
  },
  {
    key: "adminDashboard",
    url: "/admin/dashboard",
    icon: LayoutDashboardIcon,
    modes: ["provider"],
    adminOnly: true,
  },
  {
    key: "users",
    url: "/users",
    icon: UsersIcon,
    modes: ["provider"],
    temp: true,
    adminOnly: true,
  },
  {
    key: "countries",
    url: "/countries",
    icon: GlobeIcon,
    modes: ["provider"],
    temp: true,
    adminOnly: true,
  },
  {
    key: "taxes",
    url: "/taxes",
    icon: PercentIcon,
    modes: ["provider"],
    temp: true,
    adminOnly: true,
  },
  {
    key: "fees",
    url: "/fees",
    icon: CoinsIcon,
    modes: ["provider"],
    premium: true,
  },
  {
    key: "templates",
    url: "/templates",
    icon: WrenchIcon,
    modes: ["provider"],
    premium: true,
  },
  {
    key: "subscriptions",
    url: "/subscriptions",
    icon: CreditCardIcon,
    modes: ["provider"],
    adminOnly: true,
  },
  {
    key: "logs",
    url: "/logs",
    icon: ShieldIcon,
    modes: ["provider"],
    adminOnly: true,
  },
  {
    key: "emailLogs",
    url: "/email-logs",
    icon: MailIcon,
    modes: ["provider"],
    adminOnly: true,
  },
  // {
  //   key: "test",
  //   url: "/test",
  //   icon: WrenchIcon,
  //   modes: ["provider"],
  // },
];

export default function AppSidebar() {
  const { mode, setMode, isCustomer } = useMode();
  const t = useTranslations("nav");
  const [sidebarView, setSidebarView] = useState<SidebarView>("user");

  const { auth, ok } = useAuth();

  const isAdmin = useMemo(() => ok && isSuperAdmin(auth), [auth, ok]);
  const isPremium = useMemo(() => ok && canUsePaidFeatures(auth), [auth, ok]);

  const effectiveSidebarView: SidebarView = isAdmin ? sidebarView : "user";

  const visibleItems = useMemo(
    () =>
      items.filter(
        (item) =>
          (!item.modes || item.modes.includes(mode)) &&
          (!item.adminOnly || isAdmin) &&
          (!item.premium || isPremium),
      ),
    [mode, isAdmin, isPremium],
  );

  const userItems = useMemo(
    () => visibleItems.filter((item) => !item.adminOnly),
    [visibleItems],
  );

  const adminItems = useMemo(
    () => visibleItems.filter((item) => item.adminOnly),
    [visibleItems],
  );

  const shownItems = effectiveSidebarView === "admin" ? adminItems : userItems;

  const shownGroupLabel =
    effectiveSidebarView === "admin"
      ? t("adminGroupLabel")
      : t("userGroupLabel");

  return (
    <Sidebar data-mode={mode}>
      <SidebarContent className="bg-primary-foreground text-primary">
        {isAdmin && (
          <SidebarGroup>
            <SidebarGroupLabel>{t("viewSwitchLabel")}</SidebarGroupLabel>
            <SidebarGroupContent>
              <div className="grid grid-cols-2 gap-2">
                <Button
                  type="button"
                  size="sm"
                  variant={sidebarView === "user" ? "default" : "outline"}
                  onClick={() => setSidebarView("user")}
                >
                  <UserIcon />
                  <span>{t("userViewButton")}</span>
                </Button>
                <Button
                  type="button"
                  size="sm"
                  variant={sidebarView === "admin" ? "default" : "outline"}
                  onClick={() => setSidebarView("admin")}
                >
                  <ShieldIcon />
                  <span>{t("adminViewButton")}</span>
                </Button>
              </div>
            </SidebarGroupContent>
          </SidebarGroup>
        )}

        <SidebarGroup>
          <SidebarGroupLabel>{shownGroupLabel}</SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {shownItems.map((item) => (
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
        <div className="px-2 pb-1">
          <button
            type="button"
            data-slot="mode-toggle"
            data-active={isCustomer ? "true" : undefined}
            onClick={() => setMode(isCustomer ? "provider" : "customer")}
            className={cn(
              "flex w-full items-center gap-2 rounded-md px-3 py-2 text-sm transition-colors",
              isCustomer
                ? "bg-primary text-primary-foreground"
                : "hover:bg-muted",
            )}
          >
            <BuildingIcon className="h-4 w-4" />
            <span>{isCustomer ? t("modeToggle.active") : t("modeToggle.inactive")}</span>
          </button>
        </div>
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
