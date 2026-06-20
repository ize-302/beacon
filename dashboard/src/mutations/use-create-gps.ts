import { createMutation, useQueryClient } from "@tanstack/solid-query";
import type { CreateGpsRequestBody } from "~/client/api";
import { gpsApi } from "~/api/client";

export const useCreateGps = () => {
  const queryClient = useQueryClient();
  return createMutation(() => ({
    mutationFn: async (body: CreateGpsRequestBody) => {
      const res = await gpsApi.createGpsDevice(body);
      return res.data.data;
    },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["gps-devices"] }),
  }));
};
