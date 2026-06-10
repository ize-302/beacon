import { useQuery } from "@tanstack/solid-query";
import { gpsApi } from "~/api/client";

export const useGetGpss = () => {
  return useQuery(() => ({
    queryKey: ["gpss"],
    queryFn: async () => {
      const res = await gpsApi.gpsGet();
      return res.data;
    },
  }));
};
