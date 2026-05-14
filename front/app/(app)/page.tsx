import { getTranslations } from "next-intl/server";

export default async function DashboardPage() {
  const t = await getTranslations("dashboard");
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4">{t("title")}</h1>
      <p>{t("welcome")}</p>
    </div>
  );
}
