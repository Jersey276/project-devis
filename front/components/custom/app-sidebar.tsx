"use client";

import { useEffect, useMemo, useState } from "react";
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
  BarChart2Icon,
  CoinsIcon,
  ReceiptEuroIcon,
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
import { apiFetch } from "@/lib/api";
import { isSuperAdmin, type AuthContext } from "@/lib/access";

type NavKey =
  | "quote"
  | "schedule"
  | "invoices"
  | "creditNotes"
  | "clients"
  | "fees"
  | "users"
  | "countries"
  | "taxes"
  | "templates"
  | "subscriptions"
  | "analytics"
  | "logs"
  | "test";

type SidebarItem = {
  key: NavKey;
  url: string;
  icon: LucideIcon;
  // Modes in which this entry is visible. Omit to show in every mode.
  modes?: UserMode[];
  // Marker for entries that will be gated by the upcoming roles/permissions system.
  temp?: boolean;
  adminOnly?: boolean;
};

type SidebarView = "user" | "admin";

const items: SidebarItem[] = [
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
  },
  {
    key: "invoices",
    url: "/invoice",
    icon: ReceiptEuroIcon,
    modes: ["provider"],
  },
  {
    key: "creditNotes",
    url: "/credit-note",
    icon: ReceiptEuroIcon,
    modes: ["provider"],
  },
  {
    key: "clients",
    url: "/clients",
    icon: QuoteIcon,
    modes: ["provider"],
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
  },
  {
    key: "templates",
    url: "/templates",
    icon: WrenchIcon,
    modes: ["provider"],
  },
  {
    key: "subscriptions",
    url: "/subscriptions",
    icon: CreditCardIcon,
    modes: ["provider"],
    adminOnly: true,
  },
  {
    key: "analytics",
    url: "/analytics",
    icon: BarChart2Icon,
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
  const [isAdmin, setIsAdmin] = useState(false);
  const [sidebarView, setSidebarView] = useState<SidebarView>("user");

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/auth/me").then(({ ok, body }) => {
      if (cancelled) return;
      const auth = (body.auth ?? null) as AuthContext | null;
      setIsAdmin(ok && body.success === true && isSuperAdmin(auth));
    });
    return () => {
      cancelled = true;
    };
  }, []);

  const effectiveSidebarView: SidebarView = isAdmin ? sidebarView : "user";

  const visibleItems = useMemo(
    () =>
      items.filter(
        (item) =>
          (!item.modes || item.modes.includes(mode)) &&
          (!item.adminOnly || isAdmin),
      ),
    [mode, isAdmin],
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
        <UserMenu />
      </SidebarFooter>
    </Sidebar>
  );
}
