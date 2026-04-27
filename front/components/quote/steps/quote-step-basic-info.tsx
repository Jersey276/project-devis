import {
  Field,
  FieldError,
  FieldLabel,
} from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { toErrorProps } from "@/lib/api";

type QuoteStepBasicInfoProps = {
  projectName: string;
  clientId: string;
  isReadonly: boolean;
  emptyClients: string[];
  nameErrors?: string[];
  onProjectNameChange: (value: string) => void;
  onClientIdChange: (value: string) => void;
};

export default function QuoteStepBasicInfo({
  projectName,
  clientId,
  isReadonly,
  emptyClients,
  nameErrors,
  onProjectNameChange,
  onClientIdChange,
}: QuoteStepBasicInfoProps) {
  const hasNameError = !!nameErrors?.length;

  return (
    <div className="grid gap-4 md:max-w-xl">
      <Field data-invalid={hasNameError}>
        <FieldLabel htmlFor="project-name">Nom du projet</FieldLabel>
        <Input
          id="project-name"
          name="name"
          value={projectName}
          onChange={(event) => onProjectNameChange(event.target.value)}
          placeholder="Ex: Refonte site vitrine"
          disabled={isReadonly}
          aria-invalid={hasNameError}
        />
        <FieldError errors={toErrorProps(nameErrors)} />
      </Field>

      <div className="space-y-2">
        <FieldLabel>Client associé</FieldLabel>
        <Combobox
          items={emptyClients}
          value={clientId}
          onValueChange={(value) => onClientIdChange(value ?? "")}
        >
          <ComboboxInput
            placeholder="Aucun client disponible"
            disabled={isReadonly || emptyClients.length === 0}
          />
          <ComboboxContent>
            <ComboboxEmpty>Aucun client disponible.</ComboboxEmpty>
            <ComboboxList>
              {(item) => (
                <ComboboxItem key={item} value={item}>
                  {item}
                </ComboboxItem>
              )}
            </ComboboxList>
          </ComboboxContent>
        </Combobox>
      </div>
    </div>
  );
}
