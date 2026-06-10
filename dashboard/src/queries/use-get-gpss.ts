import { useQuery } from "@tanstack/solid-query";
import type { Gps } from "~/types";

const baseUrl = import.meta.env.VITE_API_BASE_URL;

export const useGetGpss = () => {
  return useQuery<Gps[]>(() => ({
    queryKey: ["gpss"],
    queryFn: async () => {
      const result = await fetch(`${baseUrl}/gps`);
      if (!result.ok) throw new Error("Failed to fetch GPS data");
      return result.json();
    },
  }));
};
