import { getTranslations } from "next-intl/server";

export default async function ForgetPasswordPage() {
  const t = await getTranslations("auth.forgetPassword");
  return <div>{t("title")}</div>;
}
