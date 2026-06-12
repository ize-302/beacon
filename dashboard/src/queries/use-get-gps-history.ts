import { createQuery } from "@tanstack/solid-query";
import type { GpsGpsHistoryResponse } from "~/client/api";
import { gpsApi } from "~/api/client";

export function useGetGpsHistory(id: () => number | null) {
  return createQuery<GpsGpsHistoryResponse>(() => ({
    queryKey: ["gps-history", id()],
    queryFn: async () => {
      const res = await gpsApi.gpsDevicesIdHistoryGet(id()!);
      return res.data;
    },
    enabled: id() !== null,
  }));
}
