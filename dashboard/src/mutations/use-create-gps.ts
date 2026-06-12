import { createMutation, useQueryClient } from "@tanstack/solid-query";
import type { GpsCreateGpsRequest } from "~/client/api";
import { gpsApi } from "~/api/client";

export const useCreateGps = () => {
  const queryClient = useQueryClient();
  return createMutation(() => ({
    mutationFn: async (body: GpsCreateGpsRequest) => {
      const res = await gpsApi.gpsDevicesPost(body);
      return res.data;
    },
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: ["gps-devices"] }),
  }));
};
