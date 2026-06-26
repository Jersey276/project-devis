"use client";

import { useEffect, useState } from "react";
import { apiFetch } from "@/lib/api";
import { isSuperAdmin, type AuthContext } from "@/lib/access";
import { Skeleton } from "@/components/ui/skeleton";
import AdminDashboard from "./admin-dashboard";
import UserDashboard from "./user-dashboard";

export default function DashboardRouter() {
  const [isAdmin, setIsAdmin] = useState<boolean | null>(null);

  useEffect(() => {
    apiFetch("/api/auth/me").then(({ ok, body }) => {
      const auth = (body.auth ?? null) as AuthContext | null;
      setIsAdmin(ok && body.success === true && isSuperAdmin(auth));
    });
  }, []);

  if (isAdmin === null) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-32 w-full rounded-lg" />
        <Skeleton className="h-64 w-full rounded-lg" />
      </div>
    );
  }

  return isAdmin ? <AdminDashboard /> : <UserDashboard />;
}
