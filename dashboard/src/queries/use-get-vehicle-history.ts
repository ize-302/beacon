import { useQuery } from "@tanstack/react-query";
import type { VehicleHistoryResponse } from "~/client/api";
import { vehiclesApi } from "~/api/client";

export function useGetVehicleHistory(id: number | null) {
  return useQuery<VehicleHistoryResponse>({
    queryKey: ["vehicle-history", id],
    queryFn: async () => {
      const res = await vehiclesApi.getVehicleHistory(id!);
      return res.data.data;
    },
    enabled: id !== null,
  });
}
