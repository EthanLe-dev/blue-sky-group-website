import type { ComponentProps } from "react";
import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";

// Adapted from the shadcn/ui registry for this app's navigation and status styles.
export function SheetContent({ children, ...props }: ComponentProps<typeof Dialog.Content>) {
  return (
    <Dialog.Portal>
      <Dialog.Overlay data-slot="sheet-overlay" className="admin-sheet-overlay fixed inset-0" />
      <Dialog.Content data-slot="sheet-content" className="admin-sheet fixed inset-y-0 left-0 overflow-y-auto" {...props} aria-describedby={undefined}>
        <div className="admin-sheet-header">
          <Dialog.Title>Điều hướng quản trị</Dialog.Title>
          <Dialog.Close type="button" className="admin-icon-button" aria-label="Đóng menu">
            <X size={20} aria-hidden="true" />
          </Dialog.Close>
        </div>
        {children}
      </Dialog.Content>
    </Dialog.Portal>
  );
}

export function Badge({ children, status }: { children: string; status: string }) {
  return <span data-slot="badge" data-status={status} className="booking-badge inline-flex items-center rounded-full font-bold">{children}</span>;
}

export function Skeleton() {
  return <div data-slot="skeleton" className="booking-skeleton rounded-xl motion-safe:animate-pulse" aria-hidden="true" />;
}
