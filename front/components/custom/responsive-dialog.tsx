import React from "react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerTitle,
  DrawerTrigger,
} from "@/components/ui/drawer";

const mobile = typeof window !== "undefined" && window.innerWidth <= 768;

type WithChildren = { children?: React.ReactNode };

// Direct aliases — no className overrides needed
const ResponsiveDialog = mobile ? Drawer : Dialog;
const ResponsiveDialogTitle = mobile ? DrawerTitle : DialogTitle;
const ResponsiveDialogTrigger = mobile ? DrawerTrigger : DialogTrigger;

// Wrappers — shared or differing classNames per breakpoint
function ResponsiveDialogClose({ children }: WithChildren) {
  return mobile ? <DrawerClose /> : <DialogClose>{children}</DialogClose>;
}

function ResponsiveDialogHeader({ children }: WithChildren) {
  const Comp = mobile ? DrawerHeader : DialogHeader;
  return <Comp className="p-4 border-b">{children}</Comp>;
}

function ResponsiveDialogDescription({ children }: WithChildren) {
  const Comp = mobile ? DrawerDescription : DialogDescription;
  return <Comp className="p-4">{children}</Comp>;
}

function ResponsiveDialogContent({ children }: WithChildren) {
  return mobile ? (
    <DrawerContent>{children}</DrawerContent>
  ) : (
    <DialogContent className="flex flex-col max-h-[90vh] gap-0 p-0">
      {children}
    </DialogContent>
  );
}

function ResponsiveDialogFooter({ children }: WithChildren) {
  return mobile ? (
    <DrawerFooter>{children}</DrawerFooter>
  ) : (
    <DialogFooter className="p-4 border-t">{children}</DialogFooter>
  );
}

function ResponsiveDialogBody({ children }: WithChildren) {
  return <div className="p-4 overflow-y-auto flex-1">{children}</div>;
}

export {
  ResponsiveDialog,
  ResponsiveDialogBody,
  ResponsiveDialogClose,
  ResponsiveDialogDescription,
  ResponsiveDialogHeader,
  ResponsiveDialogContent,
  ResponsiveDialogFooter,
  ResponsiveDialogTrigger,
  ResponsiveDialogTitle,
};
