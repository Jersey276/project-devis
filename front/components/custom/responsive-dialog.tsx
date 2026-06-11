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

function ResponsiveDialog({
  open,
  onOpenChange,
  children,
}: {
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
}) {
  return mobile ? (
    <Drawer open={open} onOpenChange={onOpenChange}>
      {children}
    </Drawer>
  ) : (
    <Dialog open={open} onOpenChange={onOpenChange}>
      {children}
    </Dialog>
  );
}

function ResponsiveDialogClose({ children }: { children?: React.ReactNode }) {
  return mobile ? <DrawerClose /> : <DialogClose>{children}</DialogClose>;
}

function ResponsiveDialogDescription({
  children,
}: {
  children: React.ReactNode;
}) {
  return mobile ? (
    <DrawerDescription className="p-4">{children}</DrawerDescription>
  ) : (
    <DialogDescription className="p-4">{children}</DialogDescription>
  );
}

function ResponsiveDialogHeader({ children }: { children: React.ReactNode }) {
  return mobile ? (
    <DrawerHeader className="p-4 border-b">{children}</DrawerHeader>
  ) : (
    <DialogHeader className="p-4 border-b">{children}</DialogHeader>
  );
}

function ResponsiveDialogContent({ children }: { children: React.ReactNode }) {
  return mobile ? (
    <DrawerContent>{children}</DrawerContent>
  ) : (
    <DialogContent className="flex flex-col max-h-[90vh] gap-0 p-0">
      {children}
    </DialogContent>
  );
}

function ResponsiveDialogFooter({ children }: { children: React.ReactNode }) {
  return mobile ? (
    <DrawerFooter>{children}</DrawerFooter>
  ) : (
    <DialogFooter className="p-4 border-t">{children}</DialogFooter>
  );
}

function ResponsiveDialogTitle({ children }: { children: React.ReactNode }) {
  return mobile ? (
    <DrawerTitle>{children}</DrawerTitle>
  ) : (
    <DialogTitle>{children}</DialogTitle>
  );
}

function ResponsiveDialogTrigger({ children }: { children: React.ReactNode }) {
  return mobile ? (
    <DrawerTrigger>{children}</DrawerTrigger>
  ) : (
    <DialogTrigger>{children}</DialogTrigger>
  );
}

function ResponsiveDialogBody({ children }: { children: React.ReactNode }) {
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
