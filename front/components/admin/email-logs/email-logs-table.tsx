"use client";

import { useTranslations } from "next-intl";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { EmailLog } from "./email-logs-dashboard";

type Props = {
  logs: EmailLog[];
};

const DATE_FORMAT = new Intl.DateTimeFormat("fr-FR", { dateStyle: "short", timeStyle: "medium" });

function statusClass(status: string): string {
  return status === "failed" ? "text-red-600 font-semibold" : "text-green-600";
}

function boolIcon(value: boolean): string {
  return value ? "✓" : "—";
}

export default function EmailLogsTable({ logs }: Props) {
  const t = useTranslations("admin.emailLogs.table");

  return (
    <div className="overflow-x-auto rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16 hidden sm:table-cell">{t("id")}</TableHead>
            <TableHead>{t("toEmail")}</TableHead>
            <TableHead className="hidden md:table-cell">{t("type")}</TableHead>
            <TableHead className="hidden lg:table-cell">{t("referenceName")}</TableHead>
            <TableHead className="w-24 text-center">{t("status")}</TableHead>
            <TableHead className="w-20 text-center hidden md:table-cell">{t("opened")}</TableHead>
            <TableHead className="w-20 text-center hidden md:table-cell">{t("clicked")}</TableHead>
            <TableHead className="w-40 hidden lg:table-cell">{t("createdAt")}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {logs.map((log) => (
            <TableRow key={log.id}>
              <TableCell className="text-muted-foreground hidden sm:table-cell">{log.id}</TableCell>
              <TableCell className="max-w-40 truncate text-sm" title={log.to_email}>
                {log.to_email}
              </TableCell>
              <TableCell className="font-mono text-xs hidden md:table-cell">{log.type}</TableCell>
              <TableCell className="max-w-40 truncate text-xs hidden lg:table-cell" title={log.reference_name ?? ""}>
                {log.reference_name || "—"}
              </TableCell>
              <TableCell className={`text-center text-sm ${statusClass(log.status)}`}>
                {log.status}
              </TableCell>
              <TableCell className="text-center text-sm text-muted-foreground hidden md:table-cell">
                {log.opened !== undefined ? boolIcon(log.opened) : "—"}
              </TableCell>
              <TableCell className="text-center text-sm text-muted-foreground hidden md:table-cell">
                {log.clicked !== undefined ? boolIcon(log.clicked) : "—"}
              </TableCell>
              <TableCell className="text-xs text-muted-foreground hidden lg:table-cell">
                {DATE_FORMAT.format(new Date(log.created_at))}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  );
}
