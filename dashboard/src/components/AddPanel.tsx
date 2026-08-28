import { createSignal, Show } from "solid-js";
import { Button } from "~/components/ui/button";
import {
  TextField,
  TextFieldInput,
  TextFieldLabel,
} from "~/components/ui/text-field";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/components/ui/select";
import {
  CreateVehicleRequestBodyVehicleTypeEnum,
  type CreateVehicleRequestBodyVehicleTypeEnum as VehicleTypeEnum,
} from "~/client/api";
import { useCreateVehicle } from "~/mutations/use-create-vehicle";
import { MdSharpDirections_car } from "solid-icons/md";

export default function AddPanel() {
  const [open, setOpen] = createSignal(false);
  const [plateNumber, setPlateNumber] = createSignal("");
  const [vehicleType, setVehicleType] = createSignal<VehicleTypeEnum | null>(
    null,
  );
  const [deviceSn, setDeviceSn] = createSignal("");

  const vehicleTypeOptions = Object.values(
    CreateVehicleRequestBodyVehicleTypeEnum,
  );
  const createVehicle = useCreateVehicle();

  // One step: the vehicle starts being tracked the moment it exists.
  const handleAddVehicle = async (e: SubmitEvent) => {
    e.preventDefault();
    const type = vehicleType();
    if (!plateNumber().trim() || !type) return;
    await createVehicle.mutateAsync({
      plate_number: plateNumber().trim(),
      vehicle_type: type,
      device_sn: deviceSn().trim() || undefined,
    });
    setPlateNumber("");
    setVehicleType(null);
    setDeviceSn("");
  };

  return (
    <div class="absolute top-4 left-4 z-10 flex items-start">
      {/* Vertical toolbar */}
      <div class="flex flex-col border bg-background shadow-sm">
        <button
          onClick={() => setOpen((prev) => !prev)}
          class={`flex flex-col items-center gap-1 px-3 py-3 text-[11px] font-medium transition-colors ${open() ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
        >
          <MdSharpDirections_car size={20} />
          Vehicle
        </button>
      </div>

      {/* Flyout panel */}
      <Show when={open()}>
        <div class="w-64 border-t border-r border-b bg-background shadow-sm">
          <div class="flex items-center justify-between px-4 py-3 border-b">
            <span class="text-sm font-semibold">Add Vehicle</span>
            <button
              onClick={() => setOpen(false)}
              class="text-muted-foreground hover:text-foreground text-base leading-none"
            >
              ✕
            </button>
          </div>

          <div class="p-4">
            <form onSubmit={handleAddVehicle} class="space-y-4">
              <TextField>
                <TextFieldLabel>Plate Number</TextFieldLabel>
                <TextFieldInput
                  placeholder="e.g. LND 123 XY"
                  value={plateNumber()}
                  onInput={(e) => setPlateNumber(e.currentTarget.value)}
                />
              </TextField>
              <div class="flex flex-col gap-1">
                <label class="text-sm font-medium leading-none">
                  Vehicle Type
                </label>
                <Select
                  options={vehicleTypeOptions}
                  value={vehicleType()}
                  onChange={setVehicleType}
                  placeholder="Select type"
                  itemComponent={(props) => (
                    <SelectItem item={props.item}>
                      {props.item.rawValue}
                    </SelectItem>
                  )}
                >
                  <SelectTrigger>
                    <SelectValue<VehicleTypeEnum>>
                      {(state) => state.selectedOption()}
                    </SelectValue>
                  </SelectTrigger>
                  <SelectContent />
                </Select>
              </div>
              <TextField>
                <TextFieldLabel>Device Serial (optional)</TextFieldLabel>
                <TextFieldInput
                  placeholder="e.g. GPS-001"
                  value={deviceSn()}
                  onInput={(e) => setDeviceSn(e.currentTarget.value)}
                />
              </TextField>
              <Button
                type="submit"
                class="w-full"
                disabled={createVehicle.isPending}
              >
                {createVehicle.isPending ? "Adding..." : "Add Vehicle"}
              </Button>
              <Show when={createVehicle.isSuccess}>
                <p class="text-xs text-green-600">
                  Vehicle added and now being tracked.
                </p>
              </Show>
              <Show when={createVehicle.isError}>
                <p class="text-xs text-destructive">
                  {(createVehicle.error as Error)?.message}
                </p>
              </Show>
            </form>
          </div>
        </div>
      </Show>
    </div>
  );
}
