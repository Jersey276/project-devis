import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type UserAddress = {
  id: string;
  label: string;
  line: string;
  city: string;
  zipCode: string;
  country: string;
};

function UserProfileInformationTab() {
  return (
    <form className="grid max-w-3xl gap-4">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="grid gap-2">
          <Label htmlFor="first_name">Prénom</Label>
          <Input id="first_name" defaultValue="John" />
        </div>
        <div className="grid gap-2">
          <Label htmlFor="last_name">Nom</Label>
          <Input id="last_name" defaultValue="Doe" />
        </div>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" defaultValue="john.doe@example.com" />
      </div>
    </form>
  );
}

function UserProfileAccountTab() {
  return (
    <form className="grid max-w-3xl gap-4">
      <div className="grid gap-2">
        <Label htmlFor="account_email">Adresse mail</Label>
        <Input
          id="account_email"
          type="email"
          defaultValue="john.doe@example.com"
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="password">Mot de passe</Label>
        <Input id="password" type="password" />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="password_confirmation">Confirmation</Label>
        <Input id="password_confirmation" type="password" />
      </div>
    </form>
  );
}

type AddressTabProps = {
  addresses: UserAddress[];
};

function UserProfileAddressesTab({ addresses }: AddressTabProps) {
  return (
    <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
      {addresses.map((address) => (
        <Card key={address.id} size="sm" className="gap-3">
          <CardHeader className="pb-0">
            <CardTitle>{address.label}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-1 text-sm">
            <p>{address.line}</p>
            <p>
              {address.zipCode} {address.city}
            </p>
            <p>{address.country}</p>
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

export {
  UserProfileInformationTab,
  UserProfileAccountTab,
  UserProfileAddressesTab,
};
