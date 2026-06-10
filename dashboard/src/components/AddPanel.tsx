import { createSignal, Show } from "solid-js";
import { Button } from "~/components/ui/button";
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "~/components/ui/sheet";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
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
import type { VehiclesVehicleResponse } from "~/client/api";
import { useGetVehicles } from "~/queries/use-get-vehicles";
import { useCreateVehicle } from "~/mutations/use-create-vehicle";
import { useCreateGps } from "~/mutations/use-create-gps";

export default function AddPanel() {
  const [plateNumber, setPlateNumber] = createSignal("");
  const [sn, setSn] = createSignal("");
  const [selectedVehicle, setSelectedVehicle] = createSignal<VehiclesVehicleResponse | null>(null);

  const vehicles = useGetVehicles();
  const createVehicle = useCreateVehicle();
  const createGps = useCreateGps();

  const handleAddVehicle = async (e: SubmitEvent) => {
    e.preventDefault();
    if (!plateNumber().trim()) return;
    await createVehicle.mutateAsync({ plate_number: plateNumber().trim() });
    setPlateNumber("");
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
    <div class="absolute top-4 left-4 z-10">
      <Sheet>
        <SheetTrigger as={Button} size="sm">
          + Add
        </SheetTrigger>
        <SheetContent position="left">
          <SheetHeader>
            <SheetTitle>Add New</SheetTitle>
          </SheetHeader>

          <Tabs defaultValue="vehicle" class="mt-4">
            <TabsList class="w-full">
              <TabsTrigger value="vehicle" class="flex-1">
                Vehicle
              </TabsTrigger>
              <TabsTrigger value="gps" class="flex-1">
                GPS Device
              </TabsTrigger>
            </TabsList>

            <TabsContent value="vehicle">
              <form onSubmit={handleAddVehicle} class="space-y-4">
                <TextField>
                  <TextFieldLabel>Plate Number</TextFieldLabel>
                  <TextFieldInput
                    placeholder="e.g. LND 123 XY"
                    value={plateNumber()}
                    onInput={(e) => setPlateNumber(e.currentTarget.value)}
                  />
                </TextField>
                <Button type="submit" class="w-full" disabled={createVehicle.isPending}>
                  {createVehicle.isPending ? "Adding..." : "Add Vehicle"}
                </Button>
                <Show when={createVehicle.isSuccess}>
                  <p class="text-sm text-green-600">Vehicle added.</p>
                </Show>
                <Show when={createVehicle.isError}>
                  <p class="text-sm text-destructive">
                    {(createVehicle.error as Error)?.message}
                  </p>
                </Show>
              </form>
            </TabsContent>

            <TabsContent value="gps">
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
                  <label class="text-sm font-medium leading-none">Vehicle</label>
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
                <Button type="submit" class="w-full" disabled={createGps.isPending}>
                  {createGps.isPending ? "Adding..." : "Add GPS Device"}
                </Button>
                <Show when={createGps.isSuccess}>
                  <p class="text-sm text-green-600">GPS device added.</p>
                </Show>
                <Show when={createGps.isError}>
                  <p class="text-sm text-destructive">
                    {(createGps.error as Error)?.message}
                  </p>
                </Show>
              </form>
            </TabsContent>
          </Tabs>
        </SheetContent>
      </Sheet>
    </div>
  );
}
