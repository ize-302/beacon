import { createMutation, useQueryClient } from "@tanstack/solid-query";
import type { VehiclesCreateVehicleRequest } from "~/client/api";
import { vehiclesApi } from "~/api/client";

export const useCreateVehicle = () => {
  const queryClient = useQueryClient();
  return createMutation(() => ({
    mutationFn: async (body: VehiclesCreateVehicleRequest) => {
      const res = await vehiclesApi.vehiclesPost(body);
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vehicles"] }),
  }));
};
