"use client";

import { useEffect, useState } from "react";
import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import LogsTable from "@/components/admin/logs/logs-table";
import { apiFetch } from "@/lib/api";
import {
  getAdminUser,
  listAdminUserInvoices,
  listAdminUserQuotes,
  listAdminUserSchedules,
} from "@/lib/services/admin-users";
import type { AdminUserAccount } from "@/components/admin/types";
import type { ActivityLog } from "@/components/admin/logs/logs-dashboard";
import type {
  BackendInvoiceSummary,
  BackendQuote,
  BackendScheduleSummary,
} from "@/types/backend";

const PAGE_SIZE = 20;

const DATE_FMT = new Intl.DateTimeFormat("fr-FR", {
  dateStyle: "medium",
  timeStyle: "short",
});
const DATE_ONLY = new Intl.DateTimeFormat("fr-FR", { dateStyle: "medium" });

function formatDate(v: string | null | undefined, fallback: string): string {
  if (!v) return fallback;
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? fallback : DATE_FMT.format(d);
}

function formatDateOnly(v: string | null | undefined, fallback: string): string {
  if (!v) return fallback;
  const d = new Date(v);
  return Number.isNaN(d.getTime()) ? fallback : DATE_ONLY.format(d);
}

function formatCents(cents: number): string {
  return new Intl.NumberFormat("fr-FR", {
    style: "currency",
    currency: "EUR",
  }).format(cents / 100);
}

// ─── Profile card ─────────────────────────────────────────────────────────────

function ProfileRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <div className="flex gap-2 py-1 text-sm">
      <span className="w-36 shrink-0 font-medium text-muted-foreground">{label}</span>
      <span className="break-all">{value}</span>
    </div>
  );
}

function ProfileCard({ user }: { user: AdminUserAccount }) {
  const t = useTranslations("admin.users.detail.profile");
  const empty = t("empty");

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
      </CardHeader>
      <CardContent className="divide-y">
        <ProfileRow label={t("userId")} value={user.user_id} />
        <ProfileRow label={t("email")} value={user.email} />
        <ProfileRow label={t("firstName")} value={user.first_name || empty} />
        <ProfileRow label={t("lastName")} value={user.last_name || empty} />
        <ProfileRow
          label={t("role")}
          value={
            <Badge variant={user.role === "admin" ? "secondary" : "outline"}>
              {user.role}
            </Badge>
          }
        />
        <ProfileRow label={t("plan")} value={user.plan || empty} />
        {user.phone && <ProfileRow label={t("phone")} value={user.phone} />}
        {user.company && <ProfileRow label={t("company")} value={user.company} />}
        {user.siren && <ProfileRow label={t("siren")} value={user.siren} />}
        {user.vat && <ProfileRow label={t("vat")} value={user.vat} />}
        <ProfileRow
          label={t("lastLogin")}
          value={formatDate(user.last_login_at, "—")}
        />
        <ProfileRow
          label={t("status")}
          value={
            user.suspended ? (
              <Badge variant="destructive">{t("statusSuspended")}</Badge>
            ) : (
              <Badge variant="outline">{t("statusActive")}</Badge>
            )
          }
        />
      </CardContent>
    </Card>
  );
}

// ─── Activity logs section ────────────────────────────────────────────────────

function ActivityLogsSection({ userId }: { userId: string }) {
  const t = useTranslations("admin.users.detail.activityLogs");
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true);
    const params = new URLSearchParams({
      user_id: userId,
      page: String(page),
      page_size: String(PAGE_SIZE),
    });
    apiFetch(`/api/logs?${params.toString()}`).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setLogs((body.logs ?? []) as ActivityLog[]);
        setTotal((body.total ?? 0) as number);
      }
      setLoading(false);
    });
    return () => { cancelled = true; };
  }, [userId, page]);

  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        {loading ? (
          <div className="space-y-2">
            {Array.from({ length: 8 }).map((_, i) => (
              <Skeleton key={i} className="h-9 w-full" />
            ))}
          </div>
        ) : logs.length === 0 ? (
          <p className="text-sm text-muted-foreground">{t("empty")}</p>
        ) : (
          <>
            <LogsTable logs={logs} />
            <div className="flex flex-wrap items-center justify-between gap-2 text-sm text-muted-foreground">
              <span>{total} résultat{total > 1 ? "s" : ""}</span>
              <div className="flex gap-2">
                <button
                  className="rounded border px-3 py-1 disabled:opacity-40"
                  disabled={page <= 1}
                  onClick={() => setPage((p) => p - 1)}
                >
                  ←
                </button>
                <span>{page} / {totalPages}</span>
                <button
                  className="rounded border px-3 py-1 disabled:opacity-40"
                  disabled={page >= totalPages}
                  onClick={() => setPage((p) => p + 1)}
                >
                  →
                </button>
              </div>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  );
}

// ─── Documents section ────────────────────────────────────────────────────────

function QuotesTab({ userId }: { userId: string }) {
  const t = useTranslations("admin.users.detail.documents.quotes");
  const tStatus = useTranslations("status.quote");
  const [quotes, setQuotes] = useState<BackendQuote[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listAdminUserQuotes(userId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) setQuotes((body.quotes ?? []) as BackendQuote[]);
      setLoading(false);
    });
    return () => { cancelled = true; };
  }, [userId]);

  if (loading) return <Skeleton className="h-32 w-full" />;
  if (quotes.length === 0) return <p className="text-sm text-muted-foreground">{t("empty")}</p>;

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.name")}</TableHead>
            <TableHead>{t("columns.state")}</TableHead>
            <TableHead className="hidden sm:table-cell">{t("columns.createdAt")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {quotes.map((q) => (
            <TableRow key={q.quote_id}>
              <TableCell className="font-medium">{q.name}</TableCell>
              <TableCell>
                <Badge variant="outline">{tStatus(q.state ?? "draft")}</Badge>
              </TableCell>
              <TableCell className="hidden sm:table-cell text-muted-foreground text-sm">
                {formatDateOnly(q.created_at, "—")}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function SchedulesTab({ userId }: { userId: string }) {
  const t = useTranslations("admin.users.detail.documents.schedules");
  const [schedules, setSchedules] = useState<BackendScheduleSummary[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listAdminUserSchedules(userId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) setSchedules((body.schedules ?? []) as BackendScheduleSummary[]);
      setLoading(false);
    });
    return () => { cancelled = true; };
  }, [userId]);

  if (loading) return <Skeleton className="h-32 w-full" />;
  if (schedules.length === 0) return <p className="text-sm text-muted-foreground">{t("empty")}</p>;

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.name")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
            <TableHead className="hidden sm:table-cell">{t("columns.quoteName")}</TableHead>
            <TableHead className="hidden md:table-cell">{t("columns.startMonth")}</TableHead>
            <TableHead className="hidden md:table-cell">{t("columns.durationMonths")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {schedules.map((s) => (
            <TableRow key={s.schedule_id}>
              <TableCell className="font-medium">{s.name}</TableCell>
              <TableCell>
                <Badge variant="outline">{s.status}</Badge>
              </TableCell>
              <TableCell className="hidden sm:table-cell text-muted-foreground text-sm">
                {s.quote_name || "—"}
              </TableCell>
              <TableCell className="hidden md:table-cell text-muted-foreground text-sm">
                {s.start_month || "—"}
              </TableCell>
              <TableCell className="hidden md:table-cell text-muted-foreground text-sm">
                {s.duration_months ? t("durationValue", { n: s.duration_months }) : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function InvoicesTab({ userId }: { userId: string }) {
  const t = useTranslations("admin.users.detail.documents.invoices");
  const [invoices, setInvoices] = useState<BackendInvoiceSummary[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    listAdminUserInvoices(userId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) setInvoices((body.invoices ?? []) as BackendInvoiceSummary[]);
      setLoading(false);
    });
    return () => { cancelled = true; };
  }, [userId]);

  if (loading) return <Skeleton className="h-32 w-full" />;
  if (invoices.length === 0) return <p className="text-sm text-muted-foreground">{t("empty")}</p>;

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t("columns.number")}</TableHead>
            <TableHead>{t("columns.status")}</TableHead>
            <TableHead className="hidden sm:table-cell">{t("columns.issuedAt")}</TableHead>
            <TableHead className="text-right">{t("columns.total")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {invoices.map((inv) => (
            <TableRow key={inv.invoice_id}>
              <TableCell className="font-medium font-mono">
                {inv.invoice_number || "—"}
              </TableCell>
              <TableCell>
                <Badge variant="outline">{inv.status}</Badge>
              </TableCell>
              <TableCell className="hidden sm:table-cell text-muted-foreground text-sm">
                {formatDateOnly(inv.issued_at, "—")}
              </TableCell>
              <TableCell className="text-right font-mono text-sm">
                {inv.total_ttc_cents ? formatCents(inv.total_ttc_cents) : "—"}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}

function DocumentsSection({ userId }: { userId: string }) {
  const t = useTranslations("admin.users.detail.documents");

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t("title")}</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="quotes">
          <TabsList className="mb-4">
            <TabsTrigger value="quotes">{t("tabs.quotes")}</TabsTrigger>
            <TabsTrigger value="schedules">{t("tabs.schedules")}</TabsTrigger>
            <TabsTrigger value="invoices">{t("tabs.invoices")}</TabsTrigger>
          </TabsList>
          <TabsContent value="quotes">
            <QuotesTab userId={userId} />
          </TabsContent>
          <TabsContent value="schedules">
            <SchedulesTab userId={userId} />
          </TabsContent>
          <TabsContent value="invoices">
            <InvoicesTab userId={userId} />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  );
}

// ─── Root component ───────────────────────────────────────────────────────────

export default function AdminUserDetail({ userId }: { userId: string }) {
  const [user, setUser] = useState<AdminUserAccount | null>(null);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getAdminUser(userId).then(({ ok, body }) => {
      if (cancelled) return;
      if (ok && body.success) {
        setUser(body.user as AdminUserAccount);
      } else {
        setNotFound(true);
      }
    });
    return () => { cancelled = true; };
  }, [userId]);

  if (notFound) {
    return (
      <Card>
        <CardContent className="py-16 text-center text-muted-foreground">
          Utilisateur introuvable.
        </CardContent>
      </Card>
    );
  }

  if (!user) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-64 w-full" />
        <Skeleton className="h-48 w-full" />
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <ProfileCard user={user} />
      <ActivityLogsSection userId={userId} />
      <DocumentsSection userId={userId} />
    </div>
  );
}
