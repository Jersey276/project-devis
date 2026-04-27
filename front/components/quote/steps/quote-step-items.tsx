import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Input } from "@/components/ui/input";
import { Skeleton } from "@/components/ui/skeleton";
import {
  CheckIcon,
  Loader2Icon,
  PlusIcon,
  TriangleAlertIcon,
  Trash2Icon,
} from "lucide-react";

export type LineSaveStatus = "idle" | "saving" | "saved" | "error";

export type QuoteItemRow = {
  lineId: string;
  name: string;
  quantity: number;
  unitPriceEuros: number;
  saveStatus: LineSaveStatus;
};

type QuoteStepItemsProps = {
  items: QuoteItemRow[];
  isReadonly: boolean;
  totalAmount: number;
  isAdding: boolean;
  onNameChange: (lineId: string, value: string) => void;
  onQuantityChange: (lineId: string, value: number) => void;
  onUnitPriceChange: (lineId: string, value: number) => void;
  onRemoveItem: (lineId: string) => void;
  onAddItem: () => void;
};

function SaveIndicator({ status }: { status: LineSaveStatus }) {
  if (status === "saving") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="saving"
        className="text-muted-foreground inline-flex items-center"
        aria-label="Enregistrement en cours"
      >
        <Loader2Icon className="size-4 animate-spin" />
      </span>
    );
  }
  if (status === "saved") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="saved"
        className="inline-flex items-center text-emerald-600"
        aria-label="Enregistré"
      >
        <CheckIcon className="size-4" />
      </span>
    );
  }
  if (status === "error") {
    return (
      <span
        data-slot="line-save-indicator"
        data-status="error"
        className="text-destructive inline-flex items-center"
        aria-label="Échec d'enregistrement"
      >
        <TriangleAlertIcon className="size-4" />
      </span>
    );
  }
  return (
    <span
      data-slot="line-save-indicator"
      data-status="idle"
      className="sr-only"
    >
      idle
    </span>
  );
}

export default function QuoteStepItems({
  items,
  isReadonly,
  totalAmount,
  isAdding,
  onNameChange,
  onQuantityChange,
  onUnitPriceChange,
  onRemoveItem,
  onAddItem,
}: QuoteStepItemsProps) {
  return (
    <div className="space-y-4">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Description</TableHead>
            <TableHead>Quantité</TableHead>
            <TableHead>Prix unitaire</TableHead>
            <TableHead>Total ligne</TableHead>
            <TableHead className="w-12">État</TableHead>
            <TableHead className="w-24">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const lineTotal = item.quantity * item.unitPriceEuros;

            return (
              <TableRow key={item.lineId} data-line-id={item.lineId}>
                <TableCell>
                  <Input
                    name="line-name"
                    value={item.name}
                    onChange={(event) =>
                      onNameChange(item.lineId, event.target.value)
                    }
                    disabled={isReadonly}
                    placeholder="Prestation"
                  />
                </TableCell>
                <TableCell>
                  <Input
                    name="line-quantity"
                    type="number"
                    min={0}
                    value={item.quantity}
                    onChange={(event) =>
                      onQuantityChange(item.lineId, Number(event.target.value))
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell>
                  <Input
                    name="line-unit-price"
                    type="number"
                    min={0}
                    step="0.01"
                    value={item.unitPriceEuros}
                    onChange={(event) =>
                      onUnitPriceChange(
                        item.lineId,
                        Number(event.target.value),
                      )
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell>{lineTotal.toFixed(2)} €</TableCell>
                <TableCell>
                  <SaveIndicator status={item.saveStatus} />
                </TableCell>
                <TableCell>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => onRemoveItem(item.lineId)}
                    disabled={isReadonly || items.length <= 1}
                    aria-label="Supprimer la ligne"
                  >
                    <Trash2Icon className="size-4" />
                  </Button>
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      {!isReadonly && (
        <Button
          type="button"
          variant="ghost"
          className="h-auto w-full p-0"
          onClick={onAddItem}
          disabled={isAdding}
          aria-label="Ajouter une ligne"
        >
          <Skeleton className="flex h-14 w-full items-center justify-center">
            {isAdding ? (
              <Loader2Icon className="size-7 animate-spin" />
            ) : (
              <PlusIcon className="size-7" />
            )}
          </Skeleton>
        </Button>
      )}

      <div className="flex justify-end">
        <div className="rounded-md border px-4 py-2 text-sm font-medium">
          Montant total: {totalAmount.toFixed(2)} €
        </div>
      </div>
    </div>
  );
}
