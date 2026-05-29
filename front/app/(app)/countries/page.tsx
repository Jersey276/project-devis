import { getTranslations } from "next-intl/server";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import CountriesTab from "@/components/admin/countries/countries-tab";
import CountryGroupsTab from "@/components/admin/countries/country-groups-tab";
import AdminGuard from "@/components/custom/admin-guard";

export default async function CountriesPage() {
  const t = await getTranslations("admin.countries");
  return (
    <AdminGuard>
      <Card>
        <CardHeader>
          <CardTitle>{t("title")}</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="pays">
            <TabsList>
              <TabsTrigger value="pays">{t("tabs.countries")}</TabsTrigger>
              <TabsTrigger value="groups">{t("tabs.groups")}</TabsTrigger>
            </TabsList>
            <TabsContent value="pays" className="pt-4">
              <CountriesTab />
            </TabsContent>
            <TabsContent value="groups" className="pt-4">
              <CountryGroupsTab />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </AdminGuard>
  );
}
