import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";

type QuoteStepBasicInfoProps = {
  projectName: string;
  clientId: string;
  isReadonly: boolean;
  emptyClients: string[];
  onProjectNameChange: (value: string) => void;
  onClientIdChange: (value: string) => void;
};

export default function QuoteStepBasicInfo({
  projectName,
  clientId,
  isReadonly,
  emptyClients,
  onProjectNameChange,
  onClientIdChange,
}: QuoteStepBasicInfoProps) {
  return (
    <div className="grid gap-4 md:max-w-xl">
      <div className="space-y-2">
        <Label htmlFor="project-name">Nom du projet</Label>
        <Input
          id="project-name"
          value={projectName}
          onChange={(event) => onProjectNameChange(event.target.value)}
          placeholder="Ex: Refonte site vitrine"
          disabled={isReadonly}
        />
      </div>

      <div className="space-y-2">
        <Label>Client associé</Label>
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
