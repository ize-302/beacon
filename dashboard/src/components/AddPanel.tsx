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
  type VehiclesVehicleResponse,
  VehiclesVehicleType,
  type VehiclesVehicleType as VehicleTypeEnum,
} from "~/client/api";
import { useGetVehicles } from "~/queries/use-get-vehicles";
import { useCreateVehicle } from "~/mutations/use-create-vehicle";
import { useCreateGps } from "~/mutations/use-create-gps";
import { MdSharpDirections_car, MdFillSignal_wifi_4_bar } from "solid-icons/md";

type ActivePanel = "vehicle" | "gps";

export default function AddPanel() {
  const [active, setActive] = createSignal<ActivePanel | null>(null);
  const [plateNumber, setPlateNumber] = createSignal("");
  const [vehicleType, setVehicleType] = createSignal<VehicleTypeEnum | null>(
    null,
  );
  const [sn, setSn] = createSignal("");
  const [selectedVehicle, setSelectedVehicle] =
    createSignal<VehiclesVehicleResponse | null>(null);

  const vehicleTypeOptions = Object.values(VehiclesVehicleType);
  const vehicles = useGetVehicles();
  const createVehicle = useCreateVehicle();
  const createGps = useCreateGps();

  const toggle = (panel: ActivePanel) =>
    setActive((prev) => (prev === panel ? null : panel));

  const handleAddVehicle = async (e: SubmitEvent) => {
    e.preventDefault();
    const type = vehicleType();
    if (!plateNumber().trim() || !type) return;
    await createVehicle.mutateAsync({
      plate_number: plateNumber().trim(),
      vehicle_type: type,
    });
    setPlateNumber("");
    setVehicleType(null);
  };

  const handleAddGps = async (e: SubmitEvent) => {
    e.preventDefault();
    const v = selectedVehicle();
    if (!sn().trim() || !v) return;
    await createGps.mutateAsync({ sn: sn().trim(), vehicle_id: v.id! });
    setSn("");
    setSelectedVehicle(null);
  };

  return (
    <div class="absolute top-4 left-4 z-10 flex items-start">
      {/* Vertical toolbar */}
      <div class="flex flex-col border bg-background shadow-sm">
        <button
          onClick={() => toggle("vehicle")}
          class={`flex flex-col items-center gap-1 px-3 py-3 text-[11px] font-medium border-b transition-colors ${active() === "vehicle" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
        >
          <MdSharpDirections_car size={20} />
          Vehicle
        </button>
        <button
          onClick={() => toggle("gps")}
          class={`flex flex-col items-center gap-1 px-3 py-3 text-[11px] font-medium transition-colors ${active() === "gps" ? "bg-primary text-primary-foreground" : "hover:bg-muted text-muted-foreground hover:text-foreground"}`}
        >
          <MdFillSignal_wifi_4_bar size={20} />
          GPS
        </button>
      </div>

      {/* Flyout panel */}
      <Show when={active() !== null}>
        <div class="w-64 border-t border-r border-b bg-background shadow-sm">
          <div class="flex items-center justify-between px-4 py-3 border-b">
            <span class="text-sm font-semibold">
              {active() === "vehicle" ? "Add Vehicle" : "Add GPS Device"}
            </span>
            <button
              onClick={() => setActive(null)}
              class="text-muted-foreground hover:text-foreground text-base leading-none"
            >
              ✕
            </button>
          </div>

          <div class="p-4">
            <Show when={active() === "vehicle"}>
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
                <Button
                  type="submit"
                  class="w-full"
                  disabled={createVehicle.isPending}
                >
                  {createVehicle.isPending ? "Adding..." : "Add Vehicle"}
                </Button>
                <Show when={createVehicle.isSuccess}>
                  <p class="text-xs text-green-600">Vehicle added.</p>
                </Show>
                <Show when={createVehicle.isError}>
                  <p class="text-xs text-destructive">
                    {(createVehicle.error as Error)?.message}
                  </p>
                </Show>
              </form>
            </Show>

            <Show when={active() === "gps"}>
              <form onSubmit={handleAddGps} class="space-y-4">
                <TextField>
                  <TextFieldLabel>Serial Number</TextFieldLabel>
                  <TextFieldInput
                    placeholder="e.g. GPS-001"
                    value={sn()}
                    onInput={(e) => setSn(e.currentTarget.value)}
                  />
                </TextField>
                <div class="flex flex-col gap-1">
                  <label class="text-sm font-medium leading-none">
                    Vehicle
                  </label>
                  <Select
                    options={vehicles.data ?? []}
                    optionValue="id"
                    optionTextValue="plate_number"
                    value={selectedVehicle()}
                    onChange={setSelectedVehicle}
                    placeholder="Select a vehicle"
                    itemComponent={(props) => (
                      <SelectItem item={props.item}>
                        {props.item.rawValue.plate_number}
                      </SelectItem>
                    )}
                  >
                    <SelectTrigger>
                      <SelectValue<VehiclesVehicleResponse>>
                        {(state) => state.selectedOption()?.plate_number}
                      </SelectValue>
                    </SelectTrigger>
                    <SelectContent />
                  </Select>
                </div>
                <Button
                  type="submit"
                  class="w-full"
                  disabled={createGps.isPending}
                >
                  {createGps.isPending ? "Adding..." : "Add GPS Device"}
                </Button>
                <Show when={createGps.isSuccess}>
                  <p class="text-xs text-green-600">GPS device added.</p>
                </Show>
                <Show when={createGps.isError}>
                  <p class="text-xs text-destructive">
                    {(createGps.error as Error)?.message}
                  </p>
                </Show>
              </form>
            </Show>
          </div>
        </div>
      </Show>
    </div>
  );
}
