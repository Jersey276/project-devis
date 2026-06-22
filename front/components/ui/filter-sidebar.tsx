"use client";

import * as React from "react";
import { SlidersHorizontalIcon } from "lucide-react";
import { useIsMobile } from "@/hooks/use-mobile";
import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetFooter,
} from "@/components/ui/sheet";
import {
  Drawer,
  DrawerContent,
  DrawerHeader,
  DrawerTitle,
  DrawerFooter,
} from "@/components/ui/drawer";

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
  const isMobile = useIsMobile();

  const trigger = (
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
  );

  const body = (
    <div className="flex-1 overflow-y-auto px-4 py-4 space-y-5">
      {children}
    </div>
  );

  const resetButton = onReset && resetLabel ? (
    <Button
      variant="ghost"
      size="sm"
      className="w-full"
      onClick={() => { onReset(); setOpen(false); }}
    >
      {resetLabel}
    </Button>
  ) : null;

  if (isMobile) {
    return (
      <>
        {trigger}
        <Drawer open={open} onOpenChange={setOpen} direction="bottom">
          <DrawerContent>
            <DrawerHeader className="border-b px-4 py-3">
              <DrawerTitle className="text-sm font-medium">{title}</DrawerTitle>
            </DrawerHeader>
            {body}
            {resetButton && (
              <DrawerFooter className="border-t px-4 py-3">
                {resetButton}
              </DrawerFooter>
            )}
          </DrawerContent>
        </Drawer>
      </>
    );
  }

  return (
    <>
      {trigger}
      <Sheet open={open} onOpenChange={setOpen}>
        <SheetContent side="right" className="flex flex-col gap-0 p-0 sm:max-w-xs">
          <SheetHeader className="border-b px-4 py-3">
            <SheetTitle className="text-sm font-medium">{title}</SheetTitle>
          </SheetHeader>
          {body}
          {resetButton && (
            <SheetFooter className="border-t px-4 py-3">
              {resetButton}
            </SheetFooter>
          )}
        </SheetContent>
      </Sheet>
    </>
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
