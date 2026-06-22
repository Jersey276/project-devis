"use client";

import * as React from "react";
import { SlidersHorizontalIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from "@/components/ui/sheet";

type FilterSidebarProps = {
  triggerLabel: string;
  title: string;
  resetLabel?: string;
  activeCount?: number;
  onReset?: () => void;
  children: React.ReactNode;
};

export function FilterSidebar({
  triggerLabel,
  title,
  resetLabel,
  activeCount = 0,
  onReset,
  children,
}: FilterSidebarProps) {
  const [open, setOpen] = React.useState(false);

  return (
    <Sheet open={open} onOpenChange={setOpen}>
      <Button
        variant="outline"
        size="sm"
        className="relative"
        onClick={() => setOpen(true)}
      >
        <SlidersHorizontalIcon className="size-4" />
        {triggerLabel}
        {activeCount > 0 && (
          <span className="bg-primary text-primary-foreground absolute -top-1.5 -right-1.5 flex size-4 items-center justify-center rounded-full text-[10px] font-semibold leading-none">
            {activeCount}
          </span>
        )}
      </Button>
      <SheetContent side="right" className="flex flex-col gap-0 p-0 sm:max-w-xs">
        <SheetHeader className="border-b px-4 py-3">
          <SheetTitle className="text-sm font-medium">{title}</SheetTitle>
        </SheetHeader>
        <div className="flex-1 overflow-y-auto px-4 py-4 space-y-5">
          {children}
        </div>
        {onReset && resetLabel && (
          <SheetFooter className="border-t px-4 py-3">
            <Button variant="ghost" size="sm" className="w-full" onClick={onReset}>
              {resetLabel}
            </Button>
          </SheetFooter>
        )}
      </SheetContent>
    </Sheet>
  );
}

export function FilterSidebarSection({
  label,
  children,
}: {
  label: string;
  children: React.ReactNode;
}) {
  return (
    <div className="space-y-2">
      <p className="text-xs font-medium text-muted-foreground uppercase tracking-wide">{label}</p>
      {children}
    </div>
  );
}
