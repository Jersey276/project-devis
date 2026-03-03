import { AppLayout } from "@/app/layout";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import ClientForm from "@/components/user/client/client-form";

export default function CreateClientPage() {
  return (
    <AppLayout>
      <Card className="max-w-3xl">
        <CardHeader>
          <CardTitle>Créer un compte client</CardTitle>
          <CardDescription>
            Créez un compte de base avec les informations principales.
          </CardDescription>
        </CardHeader>

        <CardContent>
          <ClientForm />
        </CardContent>

        <CardFooter className="justify-end gap-2">
          <Button variant="outline" type="button">
            Annuler
          </Button>
          <Button type="submit" form="create-client-form">
            Créer le compte client
          </Button>
        </CardFooter>
      </Card>
    </AppLayout>
  );
}
