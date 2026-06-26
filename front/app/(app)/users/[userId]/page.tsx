import { getTranslations } from "next-intl/server";
import Link from "next/link";
import { ChevronLeftIcon } from "lucide-react";
import AdminGuard from "@/components/custom/admin-guard";
import AdminUserDetail from "@/components/admin/users/admin-user-detail";

type Props = {
  params: Promise<{ userId: string }>;
};

export default async function AdminUserDetailPage({ params }: Props) {
  const { userId } = await params;
  const t = await getTranslations("admin.users.detail");

  return (
    <AdminGuard>
      <div className="space-y-4">
        <Link
          href="/users"
          className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground"
        >
          <ChevronLeftIcon className="h-4 w-4" />
          {t("backLink")}
        </Link>
        <AdminUserDetail userId={userId} />
      </div>
    </AdminGuard>
  );
}
