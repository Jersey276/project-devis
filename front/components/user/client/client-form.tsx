import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import AddressForm from "@/components/address/address-form";

export default function ClientForm() {
  return (
    <form id="create-client-form" className="grid gap-6">
      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="grid gap-2">
          <Label htmlFor="first_name">Prénom</Label>
          <Input id="first_name" name="first_name" placeholder="Jean" />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="last_name">Nom</Label>
          <Input id="last_name" name="last_name" placeholder="Dupont" />
        </div>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="email">Adresse mail</Label>
        <Input
          id="email"
          name="email"
          type="email"
          placeholder="jean.dupont@email.com"
        />
      </div>

      <div className="grid gap-3 rounded-lg border p-4">
        <h3 className="text-sm font-medium">Adresse</h3>
        <AddressForm />
      </div>
    </form>
  );
}
