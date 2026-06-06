"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import UserInfoForm, {
  type UserProfile,
} from "@/components/user/profile/user-info-form";
import AddressesTable from "@/components/address/addresses-table";
import ConnectionForm from "@/components/user/profile/connection-form";
import SubscriptionTab from "@/components/user/profile/subscription-tab";
import { apiFetch } from "@/lib/api";

const PROFILE_TABS = ["information", "adresse", "compte", "abonnement"] as const;
type ProfileTab = (typeof PROFILE_TABS)[number];

export default function ProfileContent() {
  const t = useTranslations("profile");
  const tCommon = useTranslations("common");
  const [user, setUser] = useState<UserProfile | null>(null);
  const [loading, setLoading] = useState(true);
  const requestedTab = useSearchParams().get("tab");
  const defaultTab: ProfileTab =
    requestedTab && (PROFILE_TABS as readonly string[]).includes(requestedTab)
      ? (requestedTab as ProfileTab)
      : "information";

  useEffect(() => {
    let cancelled = false;
    apiFetch("/api/users/me").then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success && body.user) {
        setUser(body.user as UserProfile);
      }
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <Card>
      <CardHeader>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="space-y-1">
            <CardTitle>{t("title")}</CardTitle>
            <CardDescription>
              {user?.email ?? (loading ? tCommon("actions.loading") : "")}
            </CardDescription>
          </div>
          <Badge variant={user?.suspended ? "destructive" : "outline"}>
            {user?.suspended ? t("suspendedBadge") : t("badge")}
          </Badge>
        </div>
      </CardHeader>

      <CardContent>
        {loading || !user ? (
          <p className="text-muted-foreground text-sm">{t("loadingFull")}</p>
        ) : (
          <div className="space-y-4">
            {user.suspended ? (
              <p className="text-muted-foreground rounded-md border border-dashed px-3 py-2 text-sm">
                {t("suspendedNotice")}
              </p>
            ) : null}

            <Tabs defaultValue={defaultTab}>
              <TabsList>
                <TabsTrigger value="information">
                  {t("tabs.information")}
                </TabsTrigger>
                <TabsTrigger value="adresse">{t("tabs.addresses")}</TabsTrigger>
                <TabsTrigger value="compte">{t("tabs.connection")}</TabsTrigger>
                <TabsTrigger value="abonnement">{t("tabs.subscription")}</TabsTrigger>
              </TabsList>

              <TabsContent value="information" className="pt-4">
                <UserInfoForm
                  user={user}
                  onSaved={setUser}
                  readOnly={user.suspended}
                />
              </TabsContent>

              <TabsContent value="adresse" className="pt-4">
                <AddressesTable
                  ownerType="user"
                  ownerId={user.user_id}
                  readOnly={user.suspended}
                />
              </TabsContent>

              <TabsContent value="compte" className="pt-4">
                <ConnectionForm email={user.email} readOnly={user.suspended} />
              </TabsContent>

              <TabsContent value="abonnement" className="pt-4">
                <SubscriptionTab
                  userId={user.user_id}
                  readOnly={user.suspended}
                  email={user.email}
                  phone={user.phone}
                  name={user.company}
                />
              </TabsContent>
            </Tabs>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
