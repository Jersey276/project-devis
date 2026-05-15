import { type Country } from "@/components/address/address-form";

export type CountryGroup = {
  id: number;
  name: string;
  countries?: Country[];
};

export type { BackendTax as Tax } from "@/types/backend";
