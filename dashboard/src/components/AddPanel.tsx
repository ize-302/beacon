import { useState, type SubmitEvent } from "react";
import { Button, Card, FormGroup, HTMLSelect, InputGroup } from "@blueprintjs/core";
import { IconNames } from "@blueprintjs/icons";
import {
  CreateVehicleRequestBodyVehicleTypeEnum,
  type CreateVehicleRequestBodyVehicleTypeEnum as VehicleTypeEnum,
} from "~/client/api";
import { useCreateVehicle } from "~/mutations/use-create-vehicle";

const vehicleTypeOptions = Object.values(CreateVehicleRequestBodyVehicleTypeEnum);

export default function AddPanel({ vehicleCount }: { vehicleCount: number }) {
  const [open, setOpen] = useState(false);
  const [plateNumber, setPlateNumber] = useState("");
  const [vehicleType, setVehicleType] = useState<VehicleTypeEnum | "">("");
  const [deviceSn, setDeviceSn] = useState("");

  const createVehicle = useCreateVehicle();

  // One step: the vehicle starts being tracked the moment it exists.
  const handleAddVehicle = async (e: SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!plateNumber.trim() || !vehicleType) return;
    await createVehicle.mutateAsync({
      plate_number: plateNumber.trim(),
      vehicle_type: vehicleType,
      device_sn: deviceSn.trim() || undefined,
    });
    setPlateNumber("");
    setVehicleType("");
    setDeviceSn("");
  };

  return (
    <div className="absolute top-4 left-4 z-10 flex items-start">
      <Button
        icon={IconNames.KNOWN_VEHICLE}
        text={
          <>
            Vehicle <span className="font-semibold">{vehicleCount}</span>
          </>
        }
        active={open}
        onClick={() => setOpen((prev) => !prev)}
      />

      {open && (
        <Card className="w-64 ml-2 !p-0">
          <div className="flex items-center justify-between px-4 py-3 border-b border-border">
            <span className="text-sm font-semibold">Add Vehicle</span>
            <Button
              variant="minimal"
              size="small"
              icon={IconNames.CROSS}
              onClick={() => setOpen(false)}
            />
          </div>

          <div className="p-4">
            <form onSubmit={handleAddVehicle} className="space-y-4">
              <FormGroup label="Plate Number">
                <InputGroup
                  placeholder="e.g. LND 123 XY"
                  value={plateNumber}
                  onChange={(e) => setPlateNumber(e.target.value)}
                />
              </FormGroup>

              <FormGroup label="Vehicle Type">
                <HTMLSelect
                  fill
                  value={vehicleType}
                  onChange={(e) => setVehicleType(e.target.value as VehicleTypeEnum)}
                >
                  <option value="" disabled>
                    Select type
                  </option>
                  {vehicleTypeOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </HTMLSelect>
              </FormGroup>

              <FormGroup label="Device Serial (optional)">
                <InputGroup
                  placeholder="e.g. GPS-001"
                  value={deviceSn}
                  onChange={(e) => setDeviceSn(e.target.value)}
                />
              </FormGroup>

              <Button
                type="submit"
                fill
                intent="primary"
                loading={createVehicle.isPending}
                text={createVehicle.isPending ? "Adding..." : "Add Vehicle"}
              />

              {createVehicle.isSuccess && (
                <p className="text-xs text-green-600">
                  Vehicle added and now being tracked.
                </p>
              )}
              {createVehicle.isError && (
                <p className="text-xs text-destructive">
                  {createVehicle.error?.message}
                </p>
              )}
            </form>
          </div>
        </Card>
      )}
    </div>
  );
}
