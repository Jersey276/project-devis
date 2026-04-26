import { type Country } from "@/components/address/address-form";

export type CountryGroup = {
  id: number;
  name: string;
  countries?: Country[];
};

export type Tax = {
  id: number;
  name: string;
  rate: string;
  country_group_id: number;
};
