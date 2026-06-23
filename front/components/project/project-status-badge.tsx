"use client";

import { useTranslations } from "next-intl";
import { Badge } from "@/components/ui/badge";
import type { ProjectStatus } from "@/types/backend";

const VARIANT: Record<ProjectStatus, "default" | "secondary" | "outline"> = {
  active: "default",
  completed: "secondary",
  archived: "outline",
};

export default function ProjectStatusBadge({ status }: { status: ProjectStatus }) {
  const t = useTranslations("project.status");
  return <Badge variant={VARIANT[status]}>{t(status)}</Badge>;
}
