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
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Skeleton } from "@/components/ui/skeleton";
import { PlusIcon, Trash2Icon } from "lucide-react";
import type { QuoteItem, QuoteVat } from "@/types/backend";

type QuoteStepItemsProps = {
  items: QuoteItem[];
  isReadonly: boolean;
  totalAmount: number;
  vatOptions: QuoteVat[];
  onDescriptionChange: (id: string, value: string) => void;
  onQuantityChange: (id: string, value: number) => void;
  onUnitPriceChange: (id: string, value: number) => void;
  onVatChange: (id: string, value: QuoteVat) => void;
  onRemoveItem: (id: string) => void;
  onAddItem: () => void;
};

export default function QuoteStepItems({
  items,
  isReadonly,
  totalAmount,
  vatOptions,
  onDescriptionChange,
  onQuantityChange,
  onUnitPriceChange,
  onVatChange,
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
            <TableHead>TVA (%)</TableHead>
            <TableHead>Total ligne</TableHead>
            <TableHead className="w-24">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const lineBase = item.quantity * item.unitPrice;
            const lineTotal = lineBase + (lineBase * item.vat.rate) / 100;

            return (
              <TableRow key={item.id}>
                <TableCell>
                  <Input
                    value={item.description}
                    onChange={(event) =>
                      onDescriptionChange(item.id, event.target.value)
                    }
                    disabled={isReadonly}
                    placeholder="Prestation"
                  />
                </TableCell>
                <TableCell>
                  <Input
                    type="number"
                    min={0}
                    value={item.quantity}
                    onChange={(event) =>
                      onQuantityChange(item.id, Number(event.target.value))
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell>
                  <Input
                    type="number"
                    min={0}
                    step="0.01"
                    value={item.unitPrice}
                    onChange={(event) =>
                      onUnitPriceChange(item.id, Number(event.target.value))
                    }
                    disabled={isReadonly}
                  />
                </TableCell>
                <TableCell>
                  <Combobox
                    items={vatOptions.map((vat) => vat.name)}
                    value={item.vat.name}
                    onValueChange={(value) => {
                      const selectedVat = vatOptions.find(
                        (vatOption) => vatOption.name === value,
                      );

                      if (selectedVat) {
                        onVatChange(item.id, selectedVat);
                      }
                    }}
                  >
                    <ComboboxInput disabled={isReadonly} />
                    <ComboboxContent>
                      <ComboboxEmpty>Aucune TVA disponible.</ComboboxEmpty>
                      <ComboboxList>
                        {(vatName) => (
                          <ComboboxItem key={vatName} value={vatName}>
                            {vatName}
                          </ComboboxItem>
                        )}
                      </ComboboxList>
                    </ComboboxContent>
                  </Combobox>
                </TableCell>
                <TableCell>{lineTotal.toFixed(2)} €</TableCell>
                <TableCell>
                  <Button
                    type="button"
                    variant="ghost"
                    onClick={() => onRemoveItem(item.id)}
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
        >
          <Skeleton className="flex h-14 w-full items-center justify-center">
            <PlusIcon className="size-7" />
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
