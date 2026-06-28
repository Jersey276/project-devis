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

function labelFor(items: SelectComboboxItem[], value: string): string {
  return items.find((i) => i.value === value)?.label ?? value;
}

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
            {selected.map((v: string) => (
              <ComboboxChip key={v}>{labelFor(items, v)}</ComboboxChip>
            ))}
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

  // Track whether the current value(s) were resolvable at last mount.
  // If items arrive later (async load) and a value was unresolvable before,
  // force a remount so base-ui picks up the correct label.
  const valueIsResolvable = !props.multiple
    ? !props.value || items.some((i) => i.value === props.value)
    : true;
  const [lastResolvable, setLastResolvable] = React.useState(valueIsResolvable);

  if (incomingJson !== lastPropJson) {
    setLastPropJson(incomingJson);
    setLastResolvable(valueIsResolvable);
    setMountKey((k) => k + 1);
  } else if (!lastResolvable && valueIsResolvable) {
    // Items loaded after mount — value can now be resolved, remount to show label.
    setLastResolvable(true);
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

  const itemToStringLabel = React.useCallback(
    (value: string) => labelFor(items, value),
    [items],
  );

  return (
    <div className={cn("w-full", className)}>
    <ComboboxPrimitive.Root
      key={mountKey}
      defaultValue={frozenDefault}
      onValueChange={(next) => next != null && handleValueChange(next)}
      multiple={props.multiple ?? false}
      {...(!props.multiple && { itemToStringLabel })}
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
