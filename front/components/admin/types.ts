import { type Country } from "@/components/address/address-form";

export type CountryGroup = {
  id: number;
  name: string;
  countries?: Country[];
};

export type AdminUserRole = "user" | "admin";

export type AdminUserAccount = {
  user_id: string;
  first_name: string;
  last_name: string;
  email: string;
  role: AdminUserRole;
  plan: string;
  last_login_at: string | null;
  created_at: string;
  suspended: boolean;
  phone?: string;
  company?: string;
  siren?: string;
  vat?: string;
};

export type { BackendTax as Tax } from "@/types/backend";
