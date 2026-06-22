"use client";

import * as React from "react";
import { Combobox as ComboboxPrimitive } from "@base-ui/react";
import { cn } from "@/lib/utils";
import {
  ComboboxChips,
  ComboboxChip,
  ComboboxChipsInput,
  ComboboxInput,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";

// ─── Types ───────────────────────────────────────────────────────────────────

export type SelectComboboxItem = {
  value: string;
  label: string;
};

type SelectComboboxBaseProps = {
  items: SelectComboboxItem[];
  placeholder?: string;
  emptyLabel?: string;
  className?: string;
};

type SelectComboboxSingleProps = SelectComboboxBaseProps & {
  multiple?: false;
  value: string;
  onValueChange: (value: string) => void;
};

type SelectComboboxMultipleProps = SelectComboboxBaseProps & {
  multiple: true;
  value: string[];
  onValueChange: (value: string[]) => void;
};

export type SelectComboboxProps = SelectComboboxSingleProps | SelectComboboxMultipleProps;

// ─── Sous-composant chips — lit les valeurs depuis le contexte interne ────────

function MultipleInput({
  items,
  placeholder,
}: {
  items: SelectComboboxItem[];
  placeholder?: string;
}) {
  return (
    <ComboboxPrimitive.Value>
      {(value: string[]) => {
        const selected = value ?? [];
        return (
          <ComboboxChips>
            {selected.map((v: string) => {
              const label = items.find((i) => i.value === v)?.label ?? v;
              return <ComboboxChip key={v}>{label}</ComboboxChip>;
            })}
            <ComboboxChipsInput
              placeholder={selected.length === 0 ? placeholder : undefined}
            />
          </ComboboxChips>
        );
      }}
    </ComboboxPrimitive.Value>
  );
}

// ─── Composant principal ──────────────────────────────────────────────────────

export function SelectCombobox(props: SelectComboboxProps) {
  const { items, placeholder, emptyLabel = "Aucun résultat.", className } = props;

  const incomingJson = JSON.stringify(props.value);
  const [lastPropJson, setLastPropJson] = React.useState(incomingJson);
  const [mountKey, setMountKey] = React.useState(0);

  if (incomingJson !== lastPropJson) {
    setLastPropJson(incomingJson);
    setMountKey((k) => k + 1);
  }

  // Geler defaultValue au montage — ne doit pas changer entre deux renders.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const frozenDefault = React.useMemo(() => props.value as string & string[], [mountKey]);

  const handleValueChange = (next: string | string[]) => {
    setLastPropJson(JSON.stringify(next));
    if (props.multiple) {
      (props as SelectComboboxMultipleProps).onValueChange(next as string[]);
    } else {
      (props as SelectComboboxSingleProps).onValueChange(next as string);
    }
  };

  return (
    <div className={cn("w-full", className)}>
    <ComboboxPrimitive.Root
      key={mountKey}
      defaultValue={frozenDefault}
      onValueChange={(next) => next != null && handleValueChange(next)}
      multiple={props.multiple ?? false}
    >
      {props.multiple ? (
        <MultipleInput items={items} placeholder={placeholder} />
      ) : (
        <ComboboxInput placeholder={placeholder} />
      )}
      <ComboboxContent>
        <ComboboxEmpty>{emptyLabel}</ComboboxEmpty>
        <ComboboxList>
          {items.map((item) => (
            <ComboboxItem key={item.value} value={item.value}>
              {item.label}
            </ComboboxItem>
          ))}
        </ComboboxList>
      </ComboboxContent>
    </ComboboxPrimitive.Root>
    </div>
  );
}
