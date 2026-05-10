"use client";

import { useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
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
import { apiFetch } from "@/lib/api";

const PROFILE_TABS = ["information", "adresse", "compte"] as const;
type ProfileTab = (typeof PROFILE_TABS)[number];

export default function ProfileContent() {
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
            <CardTitle>Mon profil</CardTitle>
            <CardDescription>
              {user?.email ?? (loading ? "Chargement…" : "")}
            </CardDescription>
          </div>
          <Badge variant="outline">Mon compte</Badge>
        </div>
      </CardHeader>

      <CardContent>
        {loading || !user ? (
          <p className="text-muted-foreground text-sm">
            Chargement de votre profil…
          </p>
        ) : (
          <Tabs defaultValue={defaultTab}>
            <TabsList>
              <TabsTrigger value="information">Information</TabsTrigger>
              <TabsTrigger value="adresse">Adresses</TabsTrigger>
              <TabsTrigger value="compte">Connexion</TabsTrigger>
            </TabsList>

            <TabsContent value="information" className="pt-4">
              <UserInfoForm user={user} onSaved={setUser} />
            </TabsContent>

            <TabsContent value="adresse" className="pt-4">
              <AddressesTable ownerType="user" ownerId={user.user_id} />
            </TabsContent>

            <TabsContent value="compte" className="pt-4">
              <ConnectionForm email={user.email} />
            </TabsContent>
          </Tabs>
        )}
      </CardContent>
    </Card>
  );
}
