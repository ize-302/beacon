import { createQuery } from "@tanstack/solid-query";
import type { VehicleHistoryResponse } from "~/client/api";
import { vehiclesApi } from "~/api/client";

export function useGetVehicleHistory(id: () => number | null) {
  return createQuery<VehicleHistoryResponse>(() => ({
    queryKey: ["vehicle-history", id()],
    queryFn: async () => {
      const res = await vehiclesApi.getVehicleHistory(id()!);
      return res.data.data;
    },
    enabled: id() !== null,
  }));
}
