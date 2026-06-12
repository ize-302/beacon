import { useQuery } from "@tanstack/solid-query";
import { gpsApi } from "~/api/client";

export const useGetGpsDevices = () => {
  return useQuery(() => ({
    queryKey: ["gps-devices"],
    queryFn: async () => {
      const res = await gpsApi.gpsDevicesGet();
      return res.data;
    },
  }));
};
