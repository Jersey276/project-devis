"use client";

import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

type Country = {
  nom: string;
  code: string;
};

const countries: Country[] = [
  { nom: "France", code: "FR" },
  { nom: "Belgique", code: "BE" },
  { nom: "Suisse", code: "CH" },
  { nom: "Canada", code: "CA" },
  { nom: "Luxembourg", code: "LU" },
];

export default function AddressForm() {
  return (
    <div className="grid gap-4">
      <div className="grid gap-2">
        <Label htmlFor="address_name">Nom</Label>
        <Input
          id="address_name"
          name="address_name"
          placeholder="Adresse principale"
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="address_line">Adresse</Label>
        <Input
          id="address_line"
          name="address_line"
          placeholder="12 Rue des Lilas"
        />
      </div>

      <div className="grid gap-2">
        <Label htmlFor="address_complement">Complément</Label>
        <Input
          id="address_complement"
          name="address_complement"
          placeholder="Bâtiment B, 3ème étage"
        />
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
        <div className="grid gap-2">
          <Label htmlFor="city">Ville</Label>
          <Input id="city" name="city" placeholder="Paris" />
        </div>

        <div className="grid gap-2">
          <Label htmlFor="postal_code">Code postale</Label>
          <Input id="postal_code" name="postal_code" placeholder="75001" />
        </div>
      </div>

      <div className="grid gap-2">
        <Label htmlFor="country">Pays</Label>
        <Combobox items={countries}>
          <ComboboxInput id="country" placeholder="Sélectionner un pays" />
          <ComboboxContent>
            <ComboboxEmpty>Aucun pays trouvé.</ComboboxEmpty>
            <ComboboxList>
              {(country) => (
                <ComboboxItem key={country.code} value={country.code}>
                  {country.nom} ({country.code})
                </ComboboxItem>
              )}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
      </div>
    </div>
  );
}
