import { useMutation, useQueryClient } from "@tanstack/react-query";
import type { CreateVehicleRequestBody } from "~/client/api";
import { vehiclesApi } from "~/api/client";

export const useCreateVehicle = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (body: CreateVehicleRequestBody) => {
      const res = await vehiclesApi.createVehicle(body);
      return res.data.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["vehicles"] }),
  });
};
