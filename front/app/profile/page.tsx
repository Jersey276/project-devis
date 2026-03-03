"use client";

import { useMemo, useState } from "react";
import { AppLayout } from "@/app/layout";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  UserProfileAccountTab,
  UserProfileAddressesTab,
  UserProfileInformationTab,
} from "@/components/user/profile/user-profile-tabs";

type ProfileTab = "information" | "adresse" | "compte";

type UserAddress = {
  id: string;
  label: string;
  line: string;
  city: string;
  zipCode: string;
  country: string;
};

const tabs: { id: ProfileTab; label: string }[] = [
  { id: "information", label: "Information" },
  { id: "adresse", label: "Adresse" },
  { id: "compte", label: "Compte" },
];

const addresses: UserAddress[] = [
  {
    id: "addr-1",
    label: "Adresse principale",
    line: "12 Rue des Lilas",
    city: "Paris",
    zipCode: "75001",
    country: "France",
  },
  {
    id: "addr-2",
    label: "Adresse facturation",
    line: "2 Avenue de la République",
    city: "Lyon",
    zipCode: "69002",
    country: "France",
  },
  {
    id: "addr-3",
    label: "Adresse livraison",
    line: "8 Rue du Port",
    city: "Bordeaux",
    zipCode: "33000",
    country: "France",
  },
];

export default function ProfilePage() {
  const [activeTab, setActiveTab] = useState<ProfileTab>("information");
  const userName = useMemo(() => "John Doe", []);

  return (
    <AppLayout>
      <Card>
        <CardHeader>
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div className="space-y-1">
              <CardTitle>Mon profil</CardTitle>
              <CardDescription>{userName}</CardDescription>
            </div>
            <Badge variant="outline">Mon compte</Badge>
          </div>

          <div className="flex flex-wrap gap-2 pt-2">
            {tabs.map((tab) => (
              <Button
                key={tab.id}
                variant={activeTab === tab.id ? "default" : "outline"}
                onClick={() => setActiveTab(tab.id)}
                type="button"
              >
                {tab.label}
              </Button>
            ))}
          </div>
        </CardHeader>

        <CardContent>
          {activeTab === "information" && <UserProfileInformationTab />}

          {activeTab === "adresse" && (
            <UserProfileAddressesTab addresses={addresses} />
          )}

          {activeTab === "compte" && <UserProfileAccountTab />}
        </CardContent>

        <CardFooter className="justify-end">
          <Button type="button">Enregistrer les modifications</Button>
        </CardFooter>
      </Card>
    </AppLayout>
  );
}
