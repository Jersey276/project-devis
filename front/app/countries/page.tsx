import { AppLayout } from "@/app/layout";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import CountriesTab from "@/components/admin/countries/countries-tab";
import CountryGroupsTab from "@/components/admin/countries/country-groups-tab";

export default function CountriesPage() {
  return (
    <AppLayout>
      <Card>
        <CardHeader>
          <CardTitle>Pays</CardTitle>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="pays">
            <TabsList>
              <TabsTrigger value="pays">Pays</TabsTrigger>
              <TabsTrigger value="groups">Groupes de pays</TabsTrigger>
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
    </AppLayout>
  );
}
