import { createMutation, useQueryClient } from "@tanstack/solid-query";
import type { CreateVehicleRequestBody } from "~/client/api";
import { vehiclesApi } from "~/api/client";

export const useCreateVehicle = () => {
  const queryClient = useQueryClient();
  return createMutation(() => ({
    mutationFn: async (body: CreateVehicleRequestBody) => {
      const res = await vehiclesApi.createVehicle(body);
      return res.data.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vehicles"] }),
  }));
};
