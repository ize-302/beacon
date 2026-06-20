import { createQuery } from "@tanstack/solid-query";
import type { GpsHistoryResponse } from "~/client/api";
import { gpsApi } from "~/api/client";

export function useGetGpsHistory(id: () => number | null) {
  return createQuery<GpsHistoryResponse>(() => ({
    queryKey: ["gps-history", id()],
    queryFn: async () => {
      const res = await gpsApi.getGpsHistory(id()!);
      return res.data.data;
    },
    enabled: id() !== null,
  }));
}
